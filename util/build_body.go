package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// baseError is an error type that all other error types embed.
type baseError struct {
	DefaultErrString string
	Info             string
}

func (e baseError) Error() string {
	e.DefaultErrString = "An error occurred while executing a Gophercloud request."
	return e.choseErrString()
}

func (e baseError) choseErrString() string {
	if e.Info != "" {
		return e.Info
	}
	return e.DefaultErrString
}

// errMissingInput is the error when input is required in a particular
// situation but not provided by the user
type errMissingInput struct {
	baseError
	Argument string
}

func (e errMissingInput) Error() string {
	e.DefaultErrString = fmt.Sprintf("Missing input for argument [%s]", e.Argument)
	return e.choseErrString()
}

// BuildRequestBody
func BuildRequestBody(opts interface{}, parent string) (map[string]interface{}, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	optsMap := make(map[string]interface{})
	if optsValue.Kind() == reflect.Struct {
		//fmt.Printf("optsValue.Kind() is a reflect.Struct: %+v\n", optsValue.Kind())
		for i := 0; i < optsValue.NumField(); i++ {
			v := optsValue.Field(i)
			f := optsType.Field(i)

			if f.Name != strings.Title(f.Name) {
				//fmt.Printf("Skipping field: %s...\n", f.Name)
				continue
			}

			//fmt.Printf("Starting on field: %s...\n", f.Name)

			zero := isZero(v)
			//fmt.Printf("v is zero?: %v\n", zero)

			// if the field has a required tag that's set to "true"
			if requiredTag := f.Tag.Get("required"); requiredTag == "true" {
				//fmt.Printf("Checking required field [%s]:\n\tv: %+v\n\tisZero:%v\n", f.Name, v.Interface(), zero)
				// if the field's value is zero, return a missing-argument error
				if zero {
					// if the field has a 'required' tag, it can't have a zero-value
					err := errMissingInput{}
					err.Argument = f.Name
					return nil, err
				}
			}

			if xorTag := f.Tag.Get("xor"); xorTag != "" {
				//fmt.Printf("Checking `xor` tag for field [%s] with value %+v:\n\txorTag: %s\n", f.Name, v, xorTag)
				xorField := optsValue.FieldByName(xorTag)
				var xorFieldIsZero bool
				if reflect.ValueOf(xorField.Interface()) == reflect.Zero(xorField.Type()) {
					xorFieldIsZero = true
				} else {
					if xorField.Kind() == reflect.Ptr {
						xorField = xorField.Elem()
					}
					xorFieldIsZero = isZero(xorField)
				}
				if !(zero != xorFieldIsZero) {
					err := errMissingInput{}
					err.Argument = fmt.Sprintf("%s/%s", f.Name, xorTag)
					err.Info = fmt.Sprintf("Exactly one of %s and %s must be provided", f.Name, xorTag)
					return nil, err
				}
			}

			if orTag := f.Tag.Get("or"); orTag != "" {
				//fmt.Printf("Checking `or` tag for field with:\n\tname: %+v\n\torTag:%s\n", f.Name, orTag)
				//fmt.Printf("field is zero?: %v\n", zero)
				if zero {
					orField := optsValue.FieldByName(orTag)
					var orFieldIsZero bool
					if reflect.ValueOf(orField.Interface()) == reflect.Zero(orField.Type()) {
						orFieldIsZero = true
					} else {
						if orField.Kind() == reflect.Ptr {
							orField = orField.Elem()
						}
						orFieldIsZero = isZero(orField)
					}
					if orFieldIsZero {
						err := errMissingInput{}
						err.Argument = fmt.Sprintf("%s/%s", f.Name, orTag)
						err.Info = fmt.Sprintf("At least one of %s and %s must be provided", f.Name, orTag)
						return nil, err
					}
				}
			}

			jsonTag := f.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			if v.Kind() == reflect.Slice || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice) {
				sliceValue := v
				if sliceValue.Kind() == reflect.Ptr {
					sliceValue = sliceValue.Elem()
				}

				for i := 0; i < sliceValue.Len(); i++ {
					element := sliceValue.Index(i)
					if element.Kind() == reflect.Struct || (element.Kind() == reflect.Ptr && element.Elem().Kind() == reflect.Struct) {
						_, err := BuildRequestBody(element.Interface(), "")
						if err != nil {
							return nil, err
						}
					}
				}
			}
			if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
				if zero {
					//fmt.Printf("value before change: %+v\n", optsValue.Field(i))
					if jsonTag != "" {
						jsonTagPieces := strings.Split(jsonTag, ",")
						if len(jsonTagPieces) > 1 && jsonTagPieces[1] == "omitempty" {
							if v.CanSet() {
								if !v.IsNil() {
									if v.Kind() == reflect.Ptr {
										v.Set(reflect.Zero(v.Type()))
									}
								}
								//fmt.Printf("value after change: %+v\n", optsValue.Field(i))
							}
						}
					}
					continue
				}

				//fmt.Printf("Calling BuildRequestBody with:\n\tv: %+v\n\tf.Name:%s\n", v.Interface(), f.Name)
				_, err := BuildRequestBody(v.Interface(), f.Name)
				if err != nil {
					return nil, err
				}
			}
		}

		//fmt.Printf("opts: %+v \n", opts)

		b, err := json.Marshal(opts)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("string(b): %s\n", string(b))

		err = json.Unmarshal(b, &optsMap)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("optsMap: %+v\n", optsMap)

		if parent != "" {
			optsMap = map[string]interface{}{parent: optsMap}
		}
		//fmt.Printf("optsMap after parent added: %+v\n", optsMap)
		return optsMap, nil
	}
	// Return an error if the underlying type of 'opts' isn't a struct.
	return nil, fmt.Errorf("Options type is not a struct.")
}

func isZero(v reflect.Value) bool {
	//fmt.Printf("\n\nchecking isZero for value: %+v\n", v)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return false
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		var t time.Time

		if v.Type() == reflect.TypeOf(t) {
			if v.Interface().(time.Time).IsZero() {
				return true
			}
			return false
		}
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	//fmt.Printf("zero type for value: %+v\n\n\n", z)
	return v.Interface() == z.Interface()
}
