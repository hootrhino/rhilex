package xmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExprRuntime(t *testing.T) {
	exprRuntime := NewExprRuntime()

	t.Run("AddExpression", func(t *testing.T) {
		err := exprRuntime.AddExpression("sumCheck", "a + b > 15")
		assert.NoError(t, err, "Adding expression should not fail")
	})

	t.Run("UpdateExpression", func(t *testing.T) {
		err := exprRuntime.UpdateExpression("sumCheck", "a + b > 20")
		assert.NoError(t, err, "Updating expression should not fail")
	})

	t.Run("GetExpression", func(t *testing.T) {
		expression, err := exprRuntime.GetExpression("sumCheck")
		assert.NoError(t, err, "Getting expression should not fail")
		assert.Equal(t, "a + b > 20", expression)
	})

	t.Run("EvaluateExpression", func(t *testing.T) {
		data := map[string]any{"a": 10, "b": 15}
		result := exprRuntime.Evaluate("sumCheck", data)
		assert.True(t, result.Success)
		assert.Equal(t, true, result.Result)
	})

	t.Run("ListExpressions", func(t *testing.T) {
		expressions := exprRuntime.ListExpressions()
		assert.Contains(t, expressions, "sumCheck")
	})

	t.Run("RemoveExpression", func(t *testing.T) {
		err := exprRuntime.RemoveExpression("sumCheck")
		assert.NoError(t, err, "Removing expression should not fail")
	})
}
