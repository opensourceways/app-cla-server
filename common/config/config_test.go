package config

import (
	"errors"
	"testing"
)

type testConfig struct {
	val       int
	validated bool
	defaulted bool
}

func (c *testConfig) Validate() error {
	c.validated = true
	return nil
}

func (c *testConfig) SetDefault() {
	c.defaulted = true
}

type testConfigWithError struct{}

func (c *testConfigWithError) Validate() error {
	return errors.New("validation error")
}

type testItems struct {
	items []interface{}
}

func (c *testItems) ConfigItems() []interface{} {
	return c.items
}

type testConfigAll struct {
	val   int
	child *testConfigInner
}

func (c *testConfigAll) Validate() error { return nil }
func (c *testConfigAll) SetDefault()      { c.val = 42 }

func (c *testConfigAll) ConfigItems() []interface{} {
	if c.child != nil {
		return []interface{}{c.child}
	}
	return nil
}

type testConfigInner struct {
	val int
}

func (c *testConfigInner) SetDefault() { c.val = 1 }

func TestSetDefault(t *testing.T) {
	cfg := &testConfig{}
	SetDefault(cfg)
	if !cfg.defaulted {
		t.Error("expected SetDefault to be called")
	}
}

func TestSetDefaultNoMethods(t *testing.T) {
	// Should not panic when config has no SetDefault
	type plain struct{ val int }
	cfg := &plain{val: 1}
	SetDefault(cfg)
	// Just shouldn't panic
}

func TestSetDefaultNested(t *testing.T) {
	cfg := &testConfigAll{child: &testConfigInner{}}
	SetDefault(cfg)
	if cfg.val != 42 {
		t.Errorf("expected 42, got %d", cfg.val)
	}
	if cfg.child.val != 1 {
		t.Errorf("expected nested value 1, got %d", cfg.child.val)
	}
}

func TestValidate(t *testing.T) {
	cfg := &testConfig{}
	err := Validate(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !cfg.validated {
		t.Error("expected Validate to be called")
	}
}

func TestValidateNoMethods(t *testing.T) {
	type plain struct{ val int }
	cfg := &plain{val: 1}
	err := Validate(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateError(t *testing.T) {
	cfg := &testConfigWithError{}
	err := Validate(cfg)
	if err == nil {
		t.Error("expected error")
	}
}

func TestSetDefaultWithItems(t *testing.T) {
	inner := &testConfig{}
	cfg := &testItems{items: []interface{}{inner}}
	SetDefault(cfg)
	if !inner.defaulted {
		t.Error("expected nested SetDefault to be called")
	}
}

func TestValidateWithItems(t *testing.T) {
	inner := &testConfig{}
	cfg := &testItems{items: []interface{}{inner}}
	err := Validate(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !inner.validated {
		t.Error("expected nested Validate to be called")
	}
}

func TestValidateWithItemsError(t *testing.T) {
	inner := &testConfigWithError{}
	cfg := &testItems{items: []interface{}{inner}}
	err := Validate(cfg)
	if err == nil {
		t.Error("expected error from nested config")
	}
}
