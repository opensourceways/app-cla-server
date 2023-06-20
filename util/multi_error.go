package util

import (
	"errors"
	"strings"
)

func MultiErrors(errs ...error) error {
	m := MultiError{}

	for i := range errs {
		m.AddError(errs[i])
	}

	return m.Err()
}

func NewMultiError() MultiError {
	return MultiError{}
}

type MultiError struct {
	es []string
}

func (e *MultiError) Add(s string) {
	if e != nil {
		e.es = append(e.es, s)
	}
}

func (e *MultiError) AddError(err error) {
	if err != nil {
		e.Add(err.Error())
	}
}

func (e *MultiError) Err() error {
	if e == nil || len(e.es) == 0 {
		return nil
	}

	return errors.New(strings.Join(e.es, ". "))
}
