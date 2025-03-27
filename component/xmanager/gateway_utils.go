package xmanager

import (
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

// MapToConfig 将 Map 转换为具体的每个资源专属的结构体配置
func MapToConfig(m map[string]any, s any) error {
	validate := validator.New()
	err := mapstructure.Decode(m, s)
	if err != nil {
		return err
	}
	return validate.Struct(s)
}

// ConfigToMap 反向转换
func ConfigToMap(s any) (map[string]any, error) {
	var m map[string]any
	err := mapstructure.Decode(s, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
