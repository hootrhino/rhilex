package rhilex

import "testing"

func TestLoadAndExecuteScript(t *testing.T) {
	runtime := NewLuaRuntime()

	// Define the Lua script that includes the Action function
	script := `
    function Action(Input)
        print("Executing Action with input: " .. Input)
        return "Processed " .. Input
    end
    `

	// Load the script into the runtime
	err := runtime.LoadScript("testScript", script)
	if err != nil {
		t.Fatalf("Error loading script: %v", err)
	}
	for i := 0; i < 100; i++ {
		{
			// Test execution with some input data
			input := "test data"
			result, err := runtime.ExecuteScript("testScript", input)
			if err != nil {
				t.Fatalf("Error executing script: %v", err)
			}

			// Validate the expected output
			expected := "Processed test data"
			if result != expected {
				t.Fatalf("Expected %s, but got %s", expected, result)
			}
		}
	}

}
