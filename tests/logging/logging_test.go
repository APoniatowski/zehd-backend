package logging_test

import (
	"os"
	"zehd-backend/internal/logging"
	"strings"
	"testing"
)

// TestLogIt Logit testing, to check if logs will be written/printed
func TestLogIt(t *testing.T) {
	logFunction := "TestFunction"
	logOutput := "TestOutput"
	message := "TestMessage"

	// redirect output to discard
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	logging.LogIt(logFunction, logOutput, message)

	// restore stdout
	w.Close()
	os.Stdout = old

	// check that the log file was created and contains the expected log message
	path := os.Getenv("HOME") + "/log/backend.log"
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("Error opening log file: %v", err)
	}
	defer file.Close()

	// read the file content
	fileContent := make([]byte, 100)
	_, err = file.Read(fileContent)
	if err != nil {
		t.Errorf("Error reading log file: %v", err)
	}

	// check that the log message is present in the file
	expectedMessage := logFunction + " [ " + logOutput + " ] ==> " + message
	if !strings.Contains(string(fileContent), expectedMessage) {
		t.Errorf("Log message not found in file. Expected: %s, Found: %s", expectedMessage, string(fileContent))
	}
}
