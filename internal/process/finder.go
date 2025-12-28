package process

import (
	"fmt"
	"strings"
	"syscall"
	"time"
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

// GetProcessCreationTime returns the creation time of a process by PID
func GetProcessCreationTime(pid uint32) (time.Time, error) {
	// Open the process with QUERY_INFORMATION permission
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	var creationTime, exitTime, kernelTime, userTime windows.Filetime
	err = windows.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("GetProcessTimes failed for PID %d: %w", pid, err)
	}

	// Convert FILETIME to time.Time
	return time.Unix(0, creationTime.Nanoseconds()), nil
}

// GetProcessUptime returns how long a process has been running
func GetProcessUptime(pid uint32) (time.Duration, error) {
	creationTime, err := GetProcessCreationTime(pid)
	if err != nil {
		return 0, err
	}

	return time.Since(creationTime), nil
}

func GetProcessOldestUptimeByName(name string) (time.Duration, error) {
	processes, err := FindProcessesByName(name)
	if err != nil {
		return 0, fmt.Errorf("failed to find %s: %w", name, err)
	}

	if len(processes) == 0 {
		return 0, fmt.Errorf("%s is not running", name)
	}

	var oldestUptime time.Duration
	for i, proc := range processes {
		uptime, err := GetProcessUptime(proc.PID)
		if err != nil {
			continue
		}

		if i == 0 || uptime > oldestUptime {
			oldestUptime = uptime
		}
	}

	return oldestUptime, nil
}
