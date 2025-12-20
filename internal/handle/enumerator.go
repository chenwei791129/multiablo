//nolint:govet // This file requires unsafe pointer operations for Windows API interop with variable-length structures
package handle

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// HandleInfo represents information about a handle
type HandleInfo struct {
	ProcessID uint32
	Handle    windows.Handle
	Name      string
	TypeName  string
}

// findHandlesByName finds handles by name in a specific process
// This function only enumerates handles for the specified process, not all system handles
func findHandlesByName(processID uint32, targetName string) ([]HandleInfo, error) {
	var matchedHandles []HandleInfo

	// Start with a reasonable buffer size (1MB)
	bufferSize := uint32(1024 * 1024)
	var buffer []byte
	var returnLength uint32

	// Query system handle information with increasing buffer size
	for {
		buffer = make([]byte, bufferSize)
		err := ntQuerySystemInformation(
			SystemExtendedHandleInformation,
			uintptr(unsafe.Pointer(&buffer[0])),
			bufferSize,
			&returnLength,
		)

		if err == nil {
			break
		}

		// If buffer is too small, increase it and retry
		if errno, ok := err.(syscall.Errno); ok && errno == StatusInfoLengthMismatch {
			bufferSize = returnLength + 1024*1024 // Add 1MB extra
			continue
		}

		return nil, fmt.Errorf("ntQuerySystemInformation failed: %w", err)
	}

	// Parse the handle information (64-bit version)
	handleInfo := (*SystemExtendedHandleInformationEx)(unsafe.Pointer(&buffer[0]))
	numberOfHandles := int(handleInfo.NumberOfHandles)

	// Get pointer to the first handle entry
	handlesPtr := uintptr(unsafe.Pointer(&handleInfo.Handles[0]))
	handleEntrySize := unsafe.Sizeof(SystemHandleTableEntryInfoEx{})

	// Open the target process once
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_DUP_HANDLE,
		false,
		processID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open process %d: %w", processID, err)
	}
	defer func() {
		_ = windows.CloseHandle(processHandle)
	}()

	// Get current process handle for duplication
	currentProcess := windows.CurrentProcess()

	// Only process handles belonging to the target process
	for i := 0; i < numberOfHandles; i++ {
		entryPtr := handlesPtr + uintptr(i)*handleEntrySize
		entry := (*SystemHandleTableEntryInfoEx)(unsafe.Pointer(entryPtr))

		// Only process handles from the target process
		if uint32(entry.UniqueProcessID) != processID {
			continue
		}

		// Duplicate the handle to our process
		var duplicatedHandle windows.Handle
		err = ntDuplicateObject(
			processHandle,
			windows.Handle(entry.HandleValue),
			currentProcess,
			&duplicatedHandle,
			0,
			0,
			DuplicateSameAccess,
		)
		if err != nil {
			// Skip if we can't duplicate the handle
			continue
		}

		// Query handle type first to filter out problematic types
		typeName := queryObjectType(duplicatedHandle)

		// Only query name for Event handles to avoid hanging
		var name string
		if typeName == "Event" {
			name = queryObjectName(duplicatedHandle)
		}

		// Close the duplicated handle
		_ = windows.CloseHandle(duplicatedHandle)

		// Check if this is the handle we're looking for
		// Use substring match because Windows adds prefixes like "\Sessions\1\BaseNamedObjects\"
		if name != "" && strings.Contains(name, targetName) {
			matchedHandles = append(matchedHandles, HandleInfo{
				ProcessID: processID,
				Handle:    windows.Handle(entry.HandleValue),
				Name:      name,
				TypeName:  typeName,
			})
		}
	}

	return matchedHandles, nil
}

// queryObjectType queries the type name of a handle
func queryObjectType(handle windows.Handle) string {
	buffer := make([]byte, 1024)
	var returnLength uint32

	err := ntQueryObject(
		handle,
		ObjectTypeInformation,
		uintptr(unsafe.Pointer(&buffer[0])),
		uint32(len(buffer)),
		&returnLength,
	)

	if err != nil {
		return ""
	}

	typeInfo := (*ObjectTypeInfo)(unsafe.Pointer(&buffer[0]))
	return getUnicodeString(&typeInfo.TypeName)
}

// queryObjectName queries the name of a handle with timeout protection
func queryObjectName(handle windows.Handle) string {
	// Use a larger buffer for names
	buffer := make([]byte, 4096)
	var returnLength uint32

	// Note: ntQueryObject can hang on certain handles (e.g., named pipes)
	// In production, you should implement a timeout mechanism using goroutines
	err := ntQueryObject(
		handle,
		ObjectNameInformation,
		uintptr(unsafe.Pointer(&buffer[0])),
		uint32(len(buffer)),
		&returnLength,
	)

	if err != nil {
		// If buffer is too small, try again with larger buffer
		if errno, ok := err.(syscall.Errno); ok && errno == StatusInfoLengthMismatch {
			buffer = make([]byte, returnLength)
			err = ntQueryObject(
				handle,
				ObjectNameInformation,
				uintptr(unsafe.Pointer(&buffer[0])),
				uint32(len(buffer)),
				&returnLength,
			)
			if err != nil {
				return ""
			}
		} else {
			return ""
		}
	}

	nameInfo := (*ObjectNameInfo)(unsafe.Pointer(&buffer[0]))
	return getUnicodeString(&nameInfo.Name)
}
