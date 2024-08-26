package controller

import (
	"3mdeb/device-controller/pkg/config"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os/exec"
	"reflect"
	"strings"
)

type testControllerFactory struct {
}

func (d *testControllerFactory) ControllerFactory(init map[string]any) (reflect.Value, error) {
	test := testController{}
	err := test.init(init)
	return reflect.ValueOf(&test), err
}

func init() {
	controllerFactory["testController"] = &testControllerFactory{}
}

type testController struct {
}

func (d *testController) init(map[string]any) (err error) {
	log.Println("testController: Init")
	return nil
}

func (d *testController) Clean() (err error) {
	log.Println("testController: Clean")
	return nil
}

func (d *testController) GetTime(request map[string]any) (map[string]any, config.CustomError) {
	log.Println("testController: GetTime")
	output, err := exec.Command("date", "+%X").Output()
	if err == nil {
		return map[string]any{"time": strings.TrimSuffix(string(output[:]), "\n")}, nil
	} else {
		return map[string]any{"time": ""}, config.InternalError(err)
	}
}

func (d *testController) ImageDimensions(request map[string]any, images map[string][]io.Reader) (map[string]any, config.CustomError) {
	for key, files := range images {
		log.Printf("%s images:\n", key)
		for _, img_reader := range files {
			img, format, err := image.Decode(img_reader)
			if err != nil {
				return nil, config.WrongBodyError(err)
			}
			log.Printf("loaded %s image (%dx%d)", format, img.Bounds().Dx(), img.Bounds().Dy())
		}
	}
	return nil, nil
}
