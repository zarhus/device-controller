package controller

import (
	"3mdeb/device-controller/pkg/config"
	test_c "3mdeb/device-controller/pkg/controller/test_c"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"reflect"
	"strconv"
)

type testControllerSwigFactory struct {
}

func (d *testControllerSwigFactory) ControllerFactory(init map[string]any) (reflect.Value, error) {
	test := testControllerSwig{}
	err := test.init(init)
	return reflect.ValueOf(&test), err
}

func init() {
	controllerFactory["testControllerSwig"] = &testControllerSwigFactory{}
}

type testControllerSwig struct {
}

func (d *testControllerSwig) init(map[string]any) (err error) {
	log.Println("testController: Init")
	return nil
}

func (d *testControllerSwig) Clean() (err error) {
	log.Println("testController: Clean")
	return nil
}

func (d *testControllerSwig) GetMemory(equest map[string]any) (map[string]any, config.CustomError) {
	memory := test_c.GetSystemMemory() / 1024 / 1024
	return map[string]any{"memory": string(strconv.FormatUint(memory, 10)) + " MB"}, nil
}
