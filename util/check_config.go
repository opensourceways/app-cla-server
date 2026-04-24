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
	optsValue, optsType := normalizeOpts(opts)
	if optsValue.Kind() != reflect.Struct {
		return fmt.Errorf("options type is not a struct")
	}

	fieldChain := func(s string) string {
		if parent == "" {
			return s
		}
		return parent + "." + s
	}

	for i := 0; i < optsValue.NumField(); i++ {
		v := optsValue.Field(i)
		f := optsType.Field(i)

		if err := checkField(v, f, fieldChain); err != nil {
			return err
		}
	}
	return nil
}

func normalizeOpts(opts interface{}) (reflect.Value, reflect.Type) {
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

func checkField(v reflect.Value, f reflect.StructField, fieldChain func(string) string) error {
	if f.Tag.Get("json") == "-" || f.Name != strings.Title(f.Name) {
		return nil
	}

	if isSliceOrPtrSlice(v) {
		if err := checkSliceField(v, f, fieldChain); err != nil {
			return err
		}
	}

	if isStructOrPtrStruct(v) {
		if err := checkStructField(v, f, fieldChain); err != nil {
			return err
		}
	}

	if f.Tag.Get("required") == "true" && isZero(v) {
		return errMissingInput{errArgument: fieldChain(f.Name)}
	}
	return nil
}

func isSliceOrPtrSlice(v reflect.Value) bool {
	return v.Kind() == reflect.Slice || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice)
}

func isStructOrPtrStruct(v reflect.Value) bool {
	return v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct)
}

func checkSliceField(v reflect.Value, f reflect.StructField, fieldChain func(string) string) error {
	sliceValue := v
	if sliceValue.Kind() == reflect.Ptr {
		sliceValue = sliceValue.Elem()
	}

	for i := 0; i < sliceValue.Len(); i++ {
		element := sliceValue.Index(i)
		if isStructOrPtrStruct(element) {
			if err := CheckConfig(element.Interface(), fieldChain(f.Name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkStructField(v reflect.Value, f reflect.StructField, fieldChain func(string) string) error {
	return CheckConfig(v.Interface(), fieldChain(f.Name))
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
