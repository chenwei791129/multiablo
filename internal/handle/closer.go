package handle

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// CloseRemoteHandle closes a handle in a remote process
func closeRemoteHandle(processID uint32, handle windows.Handle) error {
	// Open the target process with PROCESS_DUP_HANDLE permission
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_DUP_HANDLE,
		false,
		processID,
	)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", processID, err)
	}
	defer func() {
		_ = windows.CloseHandle(processHandle)
	}()

	// Use ntDuplicateObject with DuplicateCloseSource to close the handle
	// This is the key trick: we duplicate the handle but immediately close the source
	var duplicatedHandle windows.Handle
	err = ntDuplicateObject(
		processHandle,        // Source process
		handle,               // Source handle
		0,                    // Target process (NULL means don't create duplicate)
		&duplicatedHandle,    // Target handle (will be NULL)
		0,                    // Desired access
		0,                    // Handle attributes
		DuplicateCloseSource, // Close the source handle
	)
	if err != nil {
		return fmt.Errorf("failed to close handle 0x%X in process %d: %w", handle, processID, err)
	}

	return nil
}

// CloseHandlesByName finds and closes all handles matching the given name in a process
func CloseHandlesByName(processID uint32, handleName string) (int, error) {
	// Find all handles matching the name
	handles, err := findHandlesByName(processID, handleName)
	if err != nil {
		return 0, fmt.Errorf("failed to find handles: %w", err)
	}

	if len(handles) == 0 {
		return 0, fmt.Errorf("no handles found with name: %s", handleName)
	}

	// Close each handle
	closedCount := 0
	var lastError error
	for _, h := range handles {
		err := closeRemoteHandle(h.ProcessID, h.Handle)
		if err != nil {
			lastError = err
			continue
		}
		closedCount++
	}

	if closedCount == 0 && lastError != nil {
		return 0, fmt.Errorf("failed to close any handles: %w", lastError)
	}

	return closedCount, nil
}
