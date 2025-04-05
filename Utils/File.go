package Utils

import "os"

func RemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func SaveFile(filePath string, content string) error { 
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}