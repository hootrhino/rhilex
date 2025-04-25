// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package rhilex

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
