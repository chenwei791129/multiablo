package process

import (
	"fmt"
	"os"
	"os/exec"
)

// LaunchProcess starts a new process from the given executable path
func LaunchProcess(executablePath string) error {
	// Check if the file exists
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found: %s", executablePath)
	}

	// Create command to launch the process
	cmd := exec.Command(executablePath)
	
	// Start the process in detached mode (don't wait for it to finish)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to launch %s: %w", executablePath, err)
	}

	return nil
}
