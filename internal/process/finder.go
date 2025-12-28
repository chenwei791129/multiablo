// Package process provides utilities for Windows process management,
// including process discovery, termination, and uptime tracking.
package process

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ProcessInfo represents information about a process
type ProcessInfo struct {
	PID  uint32
	Name string
}

// FindProcessesByName finds all processes with the given name
func FindProcessesByName(name string) ([]ProcessInfo, error) {
	// Create a snapshot of all processes
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer func() {
		_ = windows.CloseHandle(snapshot)
	}()

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	// Get the first process
	err = windows.Process32First(snapshot, &procEntry)
	if err != nil {
		return nil, fmt.Errorf("Process32First failed: %w", err)
	}

	var processes []ProcessInfo

	// Iterate through all processes
	for {
		// Convert the process name from [260]uint16 to string
		processName := syscall.UTF16ToString(procEntry.ExeFile[:])

		// Check if the process name matches (case-insensitive)
		if strings.EqualFold(processName, name) {
			processes = append(processes, ProcessInfo{
				PID:  procEntry.ProcessID,
				Name: processName,
			})
		}

		// Get the next process
		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			// No more processes
			break
		}
	}

	return processes, nil
}

// IsProcessRunning checks if a process with the given name is running
func IsProcessRunning(name string) (bool, error) {
	processes, err := FindProcessesByName(name)
	if err != nil {
		return false, err
	}

	return len(processes) > 0, nil
}

// GetProcessExecutablePath retrieves the full executable path of a process by PID
func GetProcessExecutablePath(pid uint32) (string, error) {
	// Open the process with QUERY_INFORMATION | VM_READ permissions
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return "", fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	// Query the full executable path
	var exePath [windows.MAX_PATH]uint16
	size := uint32(len(exePath))
	err = windows.QueryFullProcessImageName(handle, 0, &exePath[0], &size)
	if err != nil {
		return "", fmt.Errorf("QueryFullProcessImageName failed for PID %d: %w", pid, err)
	}

	return syscall.UTF16ToString(exePath[:size]), nil
}
