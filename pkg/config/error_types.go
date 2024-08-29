package config

import (
	"errors"
	"fmt"
	"net/http"
)

type CustomError interface {
	GetHttpCode() int
	Error() string
}

type CustomErrorImpl struct {
	HttpCode int
	err      error
}

func (d *CustomErrorImpl) GetHttpCode() int {
	return d.HttpCode
}

func (d *CustomErrorImpl) Error() string {
	return fmt.Sprintf("%v", d.err)
}

func NonexistentDeviceError(id int) CustomError {
	return &CustomErrorImpl{404, fmt.Errorf("nonexistent device: %d", id)}
}

func UnsupportedFunctionError(function string) CustomError {
	return &CustomErrorImpl{404, fmt.Errorf("unsupported function: %s", function)}
}

func WrongBodyError(err error) CustomError {
	return &CustomErrorImpl{
		http.StatusUnprocessableEntity,
		errors.Join(fmt.Errorf("Error in request body"), err),
	}
}

func InternalError(err error) CustomError {
	return &CustomErrorImpl{500, err}
}

func RequestError(err error) CustomError {
	return &CustomErrorImpl{400, err}
}
