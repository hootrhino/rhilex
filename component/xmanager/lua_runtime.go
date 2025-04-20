package xmanager

import (
	"fmt"
	lua "github.com/hootrhino/gopher-lua"
	"sync"
)

// ScriptState represents the execution state of a script
type ScriptState string

const (
	ScriptIdle    ScriptState = "idle"    // Script is not running
	ScriptRunning ScriptState = "running" // Script is currently executing
	ScriptDone    ScriptState = "done"    // Script execution completed
)

// LuaRuntime is an engine that supports Lua script management and execution
type LuaRuntime struct {
	compiledScripts map[string]*lua.LFunction // Store compiled Lua functions
	scripts         map[string]string         // Store raw Lua scripts
	scriptStates    map[string]ScriptState    // Track the state of each script
	luaVMs          map[string]*lua.LState    // Track individual Lua states for each script
	mu              sync.RWMutex              // Protect concurrent access
}

// NewLuaRuntime creates a new LuaRuntime instance
func NewLuaRuntime() *LuaRuntime {
	return &LuaRuntime{
		compiledScripts: make(map[string]*lua.LFunction),
		scripts:         make(map[string]string),
		scriptStates:    make(map[string]ScriptState),
		luaVMs:          make(map[string]*lua.LState),
	}
}

// LoadScript loads and compiles a Lua script into its own Lua VM
func (runtime *LuaRuntime) LoadScript(id, script string) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	// Generate a new UUID for the Lua VM for this script
	luaVM := lua.NewState()
	runtime.luaVMs[id] = luaVM

	// Load and compile the Lua script into the Lua VM
	err := luaVM.DoString(script)
	if err != nil {
		return fmt.Errorf("error loading script: %v", err)
	}

	// Ensure the Action function is defined in the Lua script
	luaAction := luaVM.GetGlobal("Action")
	if luaAction.Type() == lua.LTNil {
		return fmt.Errorf("Action function not found in script")
	}

	// Type assertion: Assert that luaAction is a *lua.LFunction (Lua function)
	actionFunc, ok := luaAction.(*lua.LFunction)
	if !ok {
		return fmt.Errorf("Action is not of the expected type *lua.LFunction")
	}

	// Store the compiled Action function for later use
	runtime.compiledScripts[id] = actionFunc
	runtime.scripts[id] = script
	runtime.scriptStates[id] = ScriptIdle

	return nil
}

// ExecuteScript executes a loaded Lua script with the provided input
func (runtime *LuaRuntime) ExecuteScript(id string, input any) (any, error) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	// Check if script is idle or running
	state, exists := runtime.scriptStates[id]
	if !exists {
		return nil, fmt.Errorf("script %s not found", id)
	}

	// If already running, return error
	if state == ScriptRunning {
		return nil, fmt.Errorf("script %s is already running", id)
	}

	// Mark as running
	runtime.scriptStates[id] = ScriptRunning

	// Get the Lua VM for the script
	luaVM, exists := runtime.luaVMs[id]
	if !exists {
		runtime.scriptStates[id] = ScriptIdle
		return nil, fmt.Errorf("Lua VM not found for script %s", id)
	}

	// Get the compiled Action function from pre-loaded scripts
	luaAction, exists := runtime.compiledScripts[id]
	if !exists {
		runtime.scriptStates[id] = ScriptIdle
		return nil, fmt.Errorf("Action function not loaded for script %s", id)
	}

	var luaArgs lua.LValue
	switch v := input.(type) {
	case string:
		luaArgs = lua.LString(v)
	case int:
		luaArgs = lua.LNumber(v)
	case float64:
		luaArgs = lua.LNumber(v)
	case bool:
		luaArgs = lua.LBool(v)
	case []byte:
		luaArgs = lua.LString(string(v))
	case map[string]interface{}:
		luaTable := luaVM.NewTable()
		for k, val := range v {
			luaTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", val)))
		}
		luaArgs = luaTable
	case []any:
		luaTable := luaVM.NewTable()
		for i, val := range v {
			luaTable.RawSetInt(i+1, lua.LString(fmt.Sprintf("%v", val)))
		}
		luaArgs = luaTable
	case nil:
		luaArgs = lua.LNil
	case lua.LValue:
		luaArgs = v
	default:
		luaArgs = lua.LString(fmt.Sprintf("%v", v))
	}
	// Call the Lua Action function with the input data
	err := luaVM.CallByParam(lua.P{
		Fn:      luaAction,
		NRet:    1, // Expecting 1 return value (processed result)
		Protect: true,
	}, luaArgs)
	if err != nil {
		runtime.scriptStates[id] = ScriptIdle
		return nil, fmt.Errorf("error calling Action function: %v", err)
	}

	// Get the result from the Lua stack
	result := luaVM.Get(-1)

	// Pop the result off the stack
	luaVM.Pop(1)

	// Set state to done after execution
	runtime.scriptStates[id] = ScriptDone
	runtime.scriptStates[id] = ScriptIdle

	return result.String(), nil
}
