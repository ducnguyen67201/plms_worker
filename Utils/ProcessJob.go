package Utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ProcessJob(jobID string, submissionID string, userCode string, methodName string, paramName []string) (string, error) {
	// Save the driver script to a temporary file
	filePath := fmt.Sprintf("./tmp/job_%s.py", jobID)

	err := SaveFile(filePath, userCode)
	if err != nil {
		return "", err
	}

	// Execute the driver script and capture the output
	output, err := ExecutePythonScript(filePath , fmt.Sprintf("{\"%s\": 5, \"%s\": 10}", paramName[0], paramName[1]))
	if err != nil {
		return "", err
	}

	// Remove the temporary file after execution
	// err = RemoveFile(filePath)
	// if err != nil {
	// 	return "", err
	// }

	return output, nil
}

func ExecutePythonScript(filePath string, input string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	fmt.Printf("🚀 Executing script: %s\n", filePath)
	cmd := exec.Command("python", filePath)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()

	// Always print the output
	fmt.Println("🔁 Script Output:\n", string(output))

	if err != nil {
		return string(output), fmt.Errorf("execution failed: %w", err)
	}

	return string(output), nil
}