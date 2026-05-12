package util

import (
	"errors"
	"testing"
)

func TestMultiErrors(t *testing.T) {
	err := MultiErrors()
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}

	err = MultiErrors(nil, nil)
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}

	err = MultiErrors(errors.New("err1"), errors.New("err2"))
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "err1. err2" {
		t.Errorf("expected 'err1. err2', got '%s'", err.Error())
	}

	err = MultiErrors(errors.New("only one"))
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "only one" {
		t.Errorf("expected 'only one', got '%s'", err.Error())
	}
}

func TestNewMultiError(t *testing.T) {
	m := NewMultiError()
	if err := m.Err(); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestMultiErrorAdd(t *testing.T) {
	m := NewMultiError()
	m.Add("error one")
	m.Add("error two")

	err := m.Err()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "error one. error two" {
		t.Errorf("expected 'error one. error two', got '%s'", err.Error())
	}
}

func TestMultiErrorAddError(t *testing.T) {
	m := NewMultiError()
	m.AddError(errors.New("err1"))
	m.AddError(errors.New("err2"))
	m.AddError(nil) // should be ignored

	err := m.Err()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "err1. err2" {
		t.Errorf("expected 'err1. err2', got '%s'", err.Error())
	}
}

func TestMultiErrorNil(t *testing.T) {
	var m *MultiError
	m.Add("test")   // should not panic
	m.AddError(errors.New("test")) // should not panic
	if err := m.Err(); err != nil {
		t.Errorf("expected nil for nil MultiError, got: %v", err)
	}
}
