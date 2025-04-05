package Utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func SaveFile(filePath string, content string) error {
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create and write the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Println("✅ File written successfully:", filePath)
	return nil
}

// func GenerateDriverScript(userCode string , methodName string , paramName[]string, parameter map[string]interface{}) string { 
// 	var builder strings.Builder 

// 	// Add import 
// 	builder.WriteString("from typing import List \n")
// 	builder.WriteString("import sys \n")
// 	builder.WriteString("import json \n")

// 	// Add user code
// 	builder.WriteString(userCode + "\n")

// 	// Main block 
// 	builder.WriteString("if __name__ == '__main__': \n")
// 	builder.WriteString("\tsolution = Solution() \n")
// 	builder.WriteString("\tdata = json.loads(sys.stdin.read()) \n")

// 	// * Build code to parse the parameters
// 	for _, param := range paramName { 
// 		builder.WriteString(fmt.Sprintf("\t%s = data[\"%s\"]\n", param, param))
// 	}

// 	// Call the Function 
// 	callParams := strings.Join(paramName, ", ")
// 	builder.WriteString(fmt.Sprintf("\tresult = solution.%s(%s)\n", methodName, callParams))
// 	builder.WriteString("\tprint(result) \n")
// 	return builder.String()
// }


func GenerateDriverScript(userCode string, methodName string, parameter map[string]interface{}) string {
	var builder strings.Builder

	// Add imports
	builder.WriteString("from typing import List\n")
	builder.WriteString("import sys\n")
	builder.WriteString("import json\n\n")

	// Add user code
	builder.WriteString(userCode + "\n\n")

	// Main block
	builder.WriteString("if __name__ == '__main__':\n")
	builder.WriteString("\tsolution = Solution()\n")
	builder.WriteString("\tdata = json.loads(sys.stdin.read())\n")

	// Get sorted parameter names to ensure consistent order
	paramKeys := make([]string, 0, len(parameter))
	for key := range parameter {
		paramKeys = append(paramKeys, key)
	}
	sort.Strings(paramKeys) // Optional: keeps output predictable

	// Extract parameters
	for _, param := range paramKeys {
		builder.WriteString(fmt.Sprintf("\t%s = data[\"%s\"]\n", param, param))
	}

	// Generate function call
	callParams := strings.Join(paramKeys, ", ")
	builder.WriteString(fmt.Sprintf("\tresult = solution.%s(%s)\n", methodName, callParams))
	builder.WriteString("\tprint(result)\n")

	return builder.String()
}
