/*
Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
*/

// Package utils provides utility functions for various purposes.
package util

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// baseError is an error type that all other error types embed.
type baseError struct {
	defaultErrString string
	errInfo          string
}

func (e baseError) choseErrString() string {
	if e.errInfo != "" {
		return e.errInfo
	}

	return e.defaultErrString
}

// errMissingInput is the error when input is required in a particular
// situation but not provided by the user
type errMissingInput struct {
	baseError

	errArgument string
}

// Error returns a string representation of the errMissingInput error.
func (e errMissingInput) Error() string {
	e.defaultErrString = fmt.Sprintf("Missing input for argument [%s]", e.errArgument)

	return e.choseErrString()
}

// CheckConfig checks if the required fields in a struct are provided.
func CheckConfig(opts interface{}, parent string) error {
	optsValue, optsType := getReflectInfo(opts)

	if !isValidStruct(optsValue, optsType) {
		return fmt.Errorf("options type is not a struct")
	}

	return validateFields(optsValue, optsType, parent)
}

func getReflectInfo(opts interface{}) (reflect.Value, reflect.Type) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	return optsValue, optsType
}

func isValidStruct(optsValue reflect.Value, optsType reflect.Type) bool {
	return optsValue.Kind() == reflect.Struct
}

func validateFields(optsValue reflect.Value, optsType reflect.Type, parent string) error {
	for i := 0; i < optsValue.NumField(); i++ {
		v := optsValue.Field(i)
		f := optsType.Field(i)

		if shouldSkipField(f) {
			continue
		}

		if err := validateField(v, f, parent); err != nil {
			return err
		}

		if isRequiredFieldEmpty(f, v, parent) {
			return errMissingInput{
				errArgument: getFieldName(f.Name, parent),
			}
		}
	}

	return nil
}

func isRequiredFieldEmpty(f reflect.StructField, v reflect.Value, parent string) bool {
	return f.Tag.Get("required") == "true" && isZero(v)
}

func shouldSkipField(f reflect.StructField) bool {
	return f.Tag.Get("json") == "-" || f.Name != strings.Title(f.Name)
}

func validateField(v reflect.Value, f reflect.StructField, parent string) error {
	fieldName := getFieldName(f.Name, parent)

	if isSliceType(v) {
		if err := validateSliceField(v, fieldName); err != nil {
			return err
		}
	}

	if isStructType(v) {
		return CheckConfig(v.Interface(), fieldName)
	}

	return nil
}

func getFieldName(fieldName, parent string) string {
	if parent == "" {
		return fieldName
	}
	return parent + "." + fieldName
}

func isSliceType(v reflect.Value) bool {
	return v.Kind() == reflect.Slice || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice)
}

func isStructType(v reflect.Value) bool {
	return v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct)
}

func validateSliceField(v reflect.Value, fieldName string) error {
	sliceValue := v
	if sliceValue.Kind() == reflect.Ptr {
		sliceValue = sliceValue.Elem()
	}

	for i := 0; i < sliceValue.Len(); i++ {
		element := sliceValue.Index(i)

		if isElementStruct(element) {
			if err := CheckConfig(element.Interface(), fieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

func isElementStruct(element reflect.Value) bool {
	return element.Kind() == reflect.Struct || (element.Kind() == reflect.Ptr && element.Elem().Kind() == reflect.Struct)
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return false

	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}

		return true

	case reflect.Struct:
		var t time.Time
		if v.Type() == reflect.TypeOf(t) {
			value, ok := v.Interface().(time.Time)

			return ok && value.IsZero()
		}

		return false

	default:
		// Compare other types directly:
		z := reflect.Zero(v.Type())
		return v.Interface() == z.Interface()
	}
}
