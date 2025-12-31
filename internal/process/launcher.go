package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	// CreateNoWindow prevents the creation of a console window for the process
	CreateNoWindow = 0x08000000
	// CreateNewProcessGroup creates a new process group, detaching from parent's console
	CreateNewProcessGroup = 0x00000200
)

// LaunchProcess starts a new process from the given executable path
func LaunchProcess(executablePath string) error {
	// Check if the file exists
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found: %s", executablePath)
	}

	// Create command to launch the process
	cmd := exec.Command(executablePath)

	// Set Windows-specific process attributes to prevent window minimization
	// and hide the child process window completely
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CreateNoWindow | CreateNewProcessGroup,
	}

	// Start the process in detached mode (don't wait for it to finish)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to launch %s: %w", executablePath, err)
	}

	return nil
}
