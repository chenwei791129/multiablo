package process

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

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

// GetProcessOldestUptimeByName returns the uptime of the oldest running process with the given name
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
