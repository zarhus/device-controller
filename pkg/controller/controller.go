package controller

import (
	"3mdeb/device-controller/pkg/config"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
)

// each new controller should add factory to this variable in init()
var controllerFactory = map[string]ControllerFactory{}

/*
each API function should be in this format:

	FunctionName(map[string]any) (map[string]any, *CustomError)

or if function accepts files (multipart endpoint) then:

	FunctionName(map[string]any, map[string][]io.Reader) (map[string]any, *CustomError)

One exception is Clean() function called when closing application
*/
var controllers = map[int]reflect.Value{}

type ControllerFactory interface {
	ControllerFactory(init map[string]any) (reflect.Value, error)
}

type Request struct {
	ID      int            `json:"id"`
	Request map[string]any `json:"request"`
}

func Init(devices []config.Device) error {
	no_error := 0
	for _, device := range devices {
		log.Printf("Init ID: %d (%s)\n", device.Id, device.Name)
		factory, ok := controllerFactory[device.Controller]
		if !ok {
			return fmt.Errorf("nonexistent controller: %s", device.Controller)
		}
		controller, err := factory.ControllerFactory(device.ControllerCfg)
		if err == nil {
			controllers[device.Id] = controller
			log.Println(controller.Elem().Type().Name(), "initialized")
		} else {
			log.Printf("%s Init failed: %v\n", device.Controller, err)
			no_error += 1
		}
	}
	if no_error == 0 {
		return nil
	} else {
		return fmt.Errorf("%d out of %d devices failed initialization", no_error, len(devices))
	}
}

func Clean() {
	for _, controller := range controllers {
		log.Printf("%s: Cleaning", controller.Elem().Type().Name())
		clean := controller.MethodByName("Clean")
		if !clean.IsValid() {
			log.Printf("%s: No Clean() function defined.\n", controller.Elem().Type().Name())
			return
		}
		clean.Call([]reflect.Value{})
	}
}

func call(endpoint config.Endpoint, json_body Request, files map[string][]io.Reader) (response map[string]any, err config.CustomError) {
	var ok bool
	controller, ok := controllers[json_body.ID]
	if !ok {
		return nil, config.NonexistentDeviceError(json_body.ID)
	}
	method := controller.MethodByName(endpoint.Function)
	if !method.IsValid() {
		return nil, config.UnsupportedFunctionError(endpoint.Function)
	}
	log.Printf("Call %s\n", endpoint.Function)
	in_arguments := []reflect.Value{reflect.ValueOf(json_body.Request)}
	if files != nil {
		in_arguments = append(in_arguments, reflect.ValueOf(files))
	}
	return_value := method.Call(in_arguments)
	if return_value[1].IsNil() {
		err = nil
	} else {
		err, ok = return_value[1].Interface().(config.CustomError)
		if !ok {
			return nil, config.InternalError(fmt.Errorf("couldn't cast error return"))
		}
	}
	response, ok = return_value[0].Interface().(map[string]any)
	if !ok {
		err = config.InternalError(
			errors.Join(err, fmt.Errorf("couldn't cast return response return")),
		)
		return nil, err
	}
	return response, err
}

func Call(endpoint config.Endpoint, json_body Request) (map[string]any, config.CustomError) {
	return call(endpoint, json_body, nil)
}

func CallMultipart(endpoint config.Endpoint, json_body Request, files map[string][]io.Reader) (map[string]any, config.CustomError) {
	return call(endpoint, json_body, files)
}
