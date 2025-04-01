package xmanager

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

// MapToConfig
func MapToConfig(m map[string]any, s any) error {
	validate := validator.New()
	err := mapstructure.Decode(m, s)
	if err != nil {
		return err
	}
	return validate.Struct(s)
}

// ConfigToMap converts a struct to a map.
// Handles cases where s is an interface{} but nil.
func ConfigToMap(s any) (map[string]any, error) {
	if s == nil || (reflect.ValueOf(s).Kind() == reflect.Ptr &&
		reflect.ValueOf(s).IsNil()) {
		return nil, fmt.Errorf("input cannot be nil")
	}

	var m map[string]any
	err := mapstructure.Decode(s, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
