package Test

import (
	"code_evaluator_worker/Utils"
	"testing"
)

func TestGenerateDriverScript(t *testing.T) {
	var sampleCode = `
Class Solution:
	def add(a, b): return a + b
`
	userCode := sampleCode
	methodName := "add"
	paramName := []string{"a", "b"}
	expectedOutput := `Class Solution:\n\tdef add(a, b): return a + b`

	driverScript := Utils.GenerateDriverScript(userCode, methodName, paramName)
	if driverScript != expectedOutput {
		t.Fatalf("TestGenerateDriverScript failed: expected\n%s, got\n%s", expectedOutput, driverScript)
	} else {
		t.Log("TestGenerateDriverScript passed")
	}
}

func TestCodeExecution(t *testing.T) {
	var sampleCode = `
class Solution:
	def add(self, a, b): return a + b
`
	userCode := sampleCode
	methodName := "add"
	paramName := []string{"a", "b"}

	driverScript := Utils.GenerateDriverScript(userCode, methodName, paramName, [nil])

	result, err  := Utils.ProcessJob("5e29a09e-49ef-4a7e-ae72-ed2639512f44", "test_submission", driverScript, methodName, paramName)
	if err != nil {
		t.Fatalf("TestCodeExecution failed: %v", err)
	}
	
	if result != "" {
		t.Fatalf("TestCodeExecution failed: expected empty result, got %s", result)
	} else {
		t.Log("TestCodeExecution passed")
	}
}