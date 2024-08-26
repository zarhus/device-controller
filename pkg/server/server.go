package server

import (
	"3mdeb/device-controller/pkg/config"
	"3mdeb/device-controller/pkg/controller"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"
	"io"

	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const max_upload_size = 1024 * 1024 * 10

func StartServer(cfg config.Config) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Timeout(10 * time.Second))
	r.Mount("/", getRoute(cfg))
	log.Println("Starting server on", "http://"+cfg.Server.ServerAddress)

	serv := http.Server{Addr: cfg.Server.ServerAddress, Handler: r}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		if err := serv.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}()
	if err := serv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP Server Error: %v", err)
	}
}

func getRoute(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	for _, endpoint := range cfg.Server.Endpoints {
		switch endpoint.Type {
		case "GET":
			r.Get(endpoint.Path, httpHandlerFactory(endpoint))
		case "POST":
			r.Post(endpoint.Path, httpHandlerFactory(endpoint))
		}
	}
	return r
}

func handleCustomError(w http.ResponseWriter, err config.CustomError) {
	log.Println(err.Error())
	http.Error(w, http.StatusText(err.GetHttpCode()), err.GetHttpCode())
}

func httpHandlerFactory(endpoint config.Endpoint) http.HandlerFunc {
	log.Printf("Adding endpoint: %s %s -> %s", endpoint.Type, endpoint.Path, endpoint.Function)
	if !endpoint.Multipart {
		return httpHandlerNonMultipart(endpoint)
	} else {
		return httpHandlerMultipart(endpoint)
	}
}

func writeResponse(w http.ResponseWriter, response map[string]any) {
	var response_json []byte
	var err error
	if response != nil {
		response_json, err = json.Marshal(response)
		if err != nil {
			handleCustomError(w, config.InternalError(err))
			return
		}
	} else {
		response_json = []byte(`"ok"`)
	}
	w.Write(response_json)
}

func httpHandlerNonMultipart(endpoint config.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleCustomError(w, config.RequestError(err))
			return
		}
		request := controller.Request{}
		err = json.Unmarshal(body, &request)
		if err != nil {
			handleCustomError(w, config.RequestError(err))
			return
		}
		response, custom_err := controller.Call(endpoint, request)
		if custom_err != nil {
			handleCustomError(w, custom_err)
			return
		}
		writeResponse(w, response)
	}
}

func httpHandlerMultipart(endpoint config.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(max_upload_size); err != nil {
			handleCustomError(w, config.RequestError(err))
			return
		}
		// Each multipart request needs to contain at least 'request' part
		request_part, ok := r.MultipartForm.File["request"]
		if !ok {
			handleCustomError(w, config.WrongBodyError(fmt.Errorf("no request part")))
			return
		}
		if len(request_part) != 1 {
			handleCustomError(w, config.WrongBodyError(fmt.Errorf("only one request allowed")))
			return
		}
		request_file, err := request_part[0].Open()
		if err != nil {
			handleCustomError(w, config.InternalError(err))
			return
		}
		request_bytes, err := io.ReadAll(request_file)
		if err != nil {
			handleCustomError(w, config.InternalError(err))
			return
		}
		request := controller.Request{}
		err = json.Unmarshal(request_bytes, &request)
		if err != nil {
			handleCustomError(w, config.RequestError(err))
			return
		}
		files := map[string][]io.Reader{}
		for key, headers := range r.MultipartForm.File {
			if key == "request" {
				continue
			}
			for _, header := range headers {
				file, err := header.Open()
				if err != nil {
					handleCustomError(w, config.InternalError(err))
					return
				}
				defer func() { file.Close() }()
				files[key] = append(files[key], file)
			}
		}

		response, custom_err := controller.CallMultipart(endpoint, request, files)
		if custom_err != nil {
			handleCustomError(w, custom_err)
			return
		}
		writeResponse(w, response)
	}
}
