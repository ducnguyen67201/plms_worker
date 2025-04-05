package Utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Signature struct {
	MethodName string   `json:"method_name"`
	ParamName  []string `json:"param_names"`
}

func ExtractPythonMethod(code string) (*Signature, error) {
	filePath := "./Utils/extract_signature.py"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	fmt.Printf("🚀 Executing script: %s\n", filePath)

	cmd := exec.Command("python", filePath)
	cmd.Stdin = strings.NewReader(code)

	output, err := cmd.CombinedOutput() 

	fmt.Println("🔁 Script Output:\n", string(output)) 

	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	var sig Signature
	if err := json.Unmarshal(output, &sig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w\nOutput: %s", err, string(output))
	}

	return &sig, nil
}