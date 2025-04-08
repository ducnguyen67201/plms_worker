package Utils

import (
	"fmt"
	"os"
	"path/filepath"
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

func decodeEscapedString(s string) string {
	s = strings.ReplaceAll(s, `\\`, `\`) // handle double backslashes first
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\'`, `'`)
	return s
}

func GenerateDriverScript(userCode string, methodName string, input string) string {
	var builder strings.Builder

	// ✅ Manual decoding of escaped characters
	decodedCode := decodeEscapedString(userCode)

	// Boilerplate
	builder.WriteString("from typing import List\n")
	builder.WriteString("import sys\n")
	builder.WriteString("import json\n\n")

	// User code
	builder.WriteString(decodedCode + "\n\n")

	// Driver block
	builder.WriteString("if __name__ == '__main__':\n")
	builder.WriteString("\tsolution = Solution()\n")
	builder.WriteString(fmt.Sprintf("\tresult = solution.%s(%s)\n", methodName, input))
	builder.WriteString("\tprint(result)\n")

	return builder.String()
}
