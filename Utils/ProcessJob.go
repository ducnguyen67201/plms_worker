package Utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ProcessJob(jobID string, submissionID string, userCode string, testCaseID string) (string, error) {
	// Save the driver script to a temporary file
	filePath := fmt.Sprintf("./tmp/job_%s_tc_%s.py", jobID, testCaseID)

	err := SaveFile(filePath, userCode)
	if err != nil {
		return "", err
	}
	output, err := ExecutePythonScript(filePath)
	if err != nil {
		return "", err
	}
	
	// Remove the temporary file after execution
	err = RemoveFile(filePath)
	if err != nil {
		return "", err
	}

	return output, nil
}

func ExecutePythonScript(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	log.Printf("🚀 Executing script: %s\n", filePath)
	cmd := exec.Command("python", filePath)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output), fmt.Errorf("execution failed: %w", err)
	}

	return string(output), nil
}

func ConvertInputToJson(input map[string]interface{}) (*string, error) {
	// Convert to JSON
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	// Convert bytes to string
	jsonString := string(jsonBytes)
	return &jsonString, nil
}