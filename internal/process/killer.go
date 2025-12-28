package process

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// KillProcessesByName terminates all processes with the given name
func KillProcessesByName(name string) (int, error) {
	processes, err := FindProcessesByName(name)
	if err != nil {
		return 0, fmt.Errorf("failed to find processes: %w", err)
	}

	if len(processes) == 0 {
		return 0, fmt.Errorf("no process found with name: %s", name)
	}

	killedCount := 0
	var lastError error

	for _, proc := range processes {
		// Open the process with PROCESS_TERMINATE permission
		handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, proc.PID)
		if err != nil {
			lastError = fmt.Errorf("failed to open process %d: %w", proc.PID, err)
			continue
		}

		// Terminate the process with exit code 1
		err = windows.TerminateProcess(handle, 1)
		_ = windows.CloseHandle(handle)

		if err != nil {
			lastError = fmt.Errorf("failed to terminate process %d: %w", proc.PID, err)
			continue
		}

		killedCount++
	}

	if killedCount == 0 && lastError != nil {
		return 0, fmt.Errorf("failed to kill any processes: %w", lastError)
	}

	return killedCount, nil
}
