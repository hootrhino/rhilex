package xmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Name  string `validate:"required"`
	Age   int    `validate:"gte=0"`
	Email string `validate:"required,email"`
}

func TestMapToConfig(t *testing.T) {
	t.Run("Valid Map to Config", func(t *testing.T) {
		input := map[string]any{
			"Name":  "John Doe",
			"Age":   30,
			"Email": "john.doe@example.com",
		}
		var config TestConfig
		err := MapToConfig(input, &config)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", config.Name)
		assert.Equal(t, 30, config.Age)
		assert.Equal(t, "john.doe@example.com", config.Email)
	})

	t.Run("Invalid Map to Config - Missing Required Field", func(t *testing.T) {
		input := map[string]any{
			"Age":   30,
			"Email": "john.doe@example.com",
		}
		var config TestConfig
		err := MapToConfig(input, &config)
		assert.Error(t, err)
	})

	t.Run("Invalid Map to Config - Invalid Field Value", func(t *testing.T) {
		input := map[string]any{
			"Name":  "John Doe",
			"Age":   -5, // Invalid age
			"Email": "john.doe@example.com",
		}
		var config TestConfig
		err := MapToConfig(input, &config)
		assert.Error(t, err)
	})
}

func TestConfigToMap(t *testing.T) {
	t.Run("Valid Config to Map", func(t *testing.T) {
		config := TestConfig{
			Name:  "Jane Doe",
			Age:   25,
			Email: "jane.doe@example.com",
		}
		result, err := ConfigToMap(config)
		assert.NoError(t, err)
		assert.Equal(t, "Jane Doe", result["Name"])
		assert.Equal(t, 25, result["Age"])
		assert.Equal(t, "jane.doe@example.com", result["Email"])
	})

	t.Run("Invalid Config to Map - Nil Input", func(t *testing.T) {
		var config *TestConfig = nil
		result, err := ConfigToMap(config)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
