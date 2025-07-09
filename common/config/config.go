/*
Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
*/

// Package config provides functionality for managing application configuration.
package config

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type configItems interface {
	ConfigItems() []interface{}
}

// SetDefault sets the default values in a configuration by calling the SetDefault method
// on the configSetDefault interface.
func SetDefault(cfg interface{}) {
	if f, ok := cfg.(configSetDefault); ok {
		f.SetDefault()
	}
	if f, ok := cfg.(configItems); ok {
		items := f.ConfigItems()
		for i := range items {
			SetDefault(items[i])
		}
	}
}

// Validate validates a configuration by calling the Validate method on the configValidate interface.
func Validate(cfg interface{}) error {
	if f, ok := cfg.(configValidate); ok {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	if f, ok := cfg.(configItems); ok {
		items := f.ConfigItems()
		for i := range items {
			if err := Validate(items[i]); err != nil {
				return err
			}
		}
	}

	return nil
}
