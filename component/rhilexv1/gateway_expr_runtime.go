package rhilex

import (
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// ExprRuntime 用于动态计算表达式，支持完整的生命周期管理
type ExprRuntime struct {
	compiledExpressions map[string]*vm.Program // 存储已编译的表达式
	expressions         map[string]string      // 存储原始表达式
	mu                  sync.RWMutex           // 保护并发访问
}

// ExprResult 表示表达式计算的结果
type ExprResult struct {
	Success bool   `json:"success"` // 是否成功
	Result  any    `json:"result"`  // 计算结果
	Error   string `json:"error"`   // 错误信息（如果有）
}

// NewExprRuntime 创建一个新的 ExprRuntime 实例
func NewExprRuntime() *ExprRuntime {
	return &ExprRuntime{
		compiledExpressions: make(map[string]*vm.Program),
		expressions:         make(map[string]string),
	}
}

// AddExpression 添加一个新的表达式并编译
// name: 表达式名称
// expression: 表达式字符串
func (er *ExprRuntime) AddExpression(name, expression string) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	// 编译表达式
	program, err := expr.Compile(expression)
	if err != nil {
		return fmt.Errorf("failed to compile expression '%s': %v", name, err)
	}

	// 存储表达式和已编译的程序
	er.expressions[name] = expression
	er.compiledExpressions[name] = program
	return nil
}

// UpdateExpression 更新已存在的表达式
// name: 表达式名称
// newExpression: 新的表达式字符串
func (er *ExprRuntime) UpdateExpression(name, newExpression string) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	// 检查表达式是否存在
	if _, exists := er.expressions[name]; !exists {
		return fmt.Errorf("expression '%s' not found", name)
	}

	// 编译新的表达式
	program, err := expr.Compile(newExpression)
	if err != nil {
		return fmt.Errorf("failed to compile new expression '%s': %v", name, err)
	}

	// 更新表达式和已编译的程序
	er.expressions[name] = newExpression
	er.compiledExpressions[name] = program
	return nil
}

// RemoveExpression 删除一个表达式
// name: 表达式名称
func (er *ExprRuntime) RemoveExpression(name string) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	// 检查表达式是否存在
	if _, exists := er.expressions[name]; !exists {
		return fmt.Errorf("expression '%s' not found", name)
	}

	// 删除表达式和已编译的程序
	delete(er.expressions, name)
	delete(er.compiledExpressions, name)
	return nil
}

// GetExpression 获取表达式的原始字符串
// name: 表达式名称
func (er *ExprRuntime) GetExpression(name string) (string, error) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	expression, exists := er.expressions[name]
	if !exists {
		return "", fmt.Errorf("expression '%s' not found", name)
	}
	return expression, nil
}

// Evaluate 使用已编译的表达式计算结果
// name: 表达式名称
// data: 输入数据（上下文）
func (er *ExprRuntime) Evaluate(name string, data map[string]any) ExprResult {
	er.mu.RLock()
	program, exists := er.compiledExpressions[name]
	er.mu.RUnlock()

	if !exists {
		return ExprResult{
			Success: false,
			Error:   fmt.Sprintf("expression '%s' not found", name),
		}
	}

	// 执行表达式
	output, err := expr.Run(program, data)
	if err != nil {
		return ExprResult{
			Success: false,
			Error:   fmt.Sprintf("failed to execute expression '%s': %v", name, err),
		}
	}

	return ExprResult{
		Success: true,
		Result:  output,
	}
}

// ListExpressions 列出所有表达式的名称
func (er *ExprRuntime) ListExpressions() []string {
	er.mu.RLock()
	defer er.mu.RUnlock()

	names := make([]string, 0, len(er.expressions))
	for name := range er.expressions {
		names = append(names, name)
	}
	return names
}
