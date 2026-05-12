package repository

import (
	"errors"
	"testing"
)

func TestErrorDuplicateCreating(t *testing.T) {
	err := NewErrorDuplicateCreating(errors.New("dup"))
	if !IsErrorDuplicateCreating(err) {
		t.Error("expected IsErrorDuplicateCreating=true")
	}
	if IsErrorResourceNotFound(err) {
		t.Error("expected IsErrorResourceNotFound=false")
	}
}

func TestErrorResourceNotFound(t *testing.T) {
	err := NewErrorResourceNotFound(errors.New("not found"))
	if !IsErrorResourceNotFound(err) {
		t.Error("expected IsErrorResourceNotFound=true")
	}
	if IsErrorConcurrentUpdating(err) {
		t.Error("expected IsErrorConcurrentUpdating=false")
	}
}

func TestErrorConcurrentUpdating(t *testing.T) {
	err := NewErrorConcurrentUpdating(errors.New("concurrent"))
	if !IsErrorConcurrentUpdating(err) {
		t.Error("expected IsErrorConcurrentUpdating=true")
	}
}

func TestIsErrorResourceNotFoundPlainError(t *testing.T) {
	if IsErrorResourceNotFound(errors.New("plain")) {
		t.Error("expected false for plain error")
	}
}

func TestErrorInterfaces(t *testing.T) {
	var _ error = NewErrorDuplicateCreating(errors.New("e"))
	var _ error = NewErrorResourceNotFound(errors.New("e"))
	var _ error = NewErrorConcurrentUpdating(errors.New("e"))
}
