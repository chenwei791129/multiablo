# Multiablo Copilot Instructions

## Project Overview

**Multiablo** is a Windows-based D2R (Diablo II: Resurrected) multi-instance launcher written in Go. It automatically closes the single-instance Event Handle (`DiabloII Check For Other Instances`) that prevents D2R from running multiple instances, enabling users to launch multiple D2R clients simultaneously.

**Key Platform Constraint**: Windows only (amd64/386), Windows API-heavy Go project using undocumented `ntdll.dll` functions.

---

## Architecture & Major Components

### 1. Process Discovery Layer (`internal/process/`)
- **finder.go**: Uses Windows toolhelp32 API to enumerate running D2R.exe processes
- Returns `ProcessInfo` slice with PID and command line
- **Key Functions**: 
  - `FindProcessesByName()` - Find processes by executable name
  - `IsProcessRunning()` - Check if process exists
  - `KillProcessesByName()` - Terminate processes by name
  - `GetProcessCreationTime()` - Get process start time via `GetProcessTimes` API
  - `GetProcessUptime()` - Calculate how long a process has been running
  - `GetProcessOldestUptimeByName()` - Find the oldest running process instance by name
- **Pattern**: Simple wrappers around Windows API with error handling

### 2. Handle Management Layer (`internal/handle/`)

#### winapi.go - Low-level Windows API Bindings
- Direct `syscall` wrappers for undocumented `ntdll.dll` functions:
  - `NtQuerySystemInformation()` - enumerate all system handles (2 info classes: 32-bit and 64-bit)
  - `NtQueryObject()` - query handle name and type information
  - `NtDuplicateObject()` - close handles in remote processes via `DuplicateCloseSource` flag
- **Critical Constants**:
  - `SystemExtendedHandleInformation = 64` (64-bit handle info, 32-bit uses 16)
  - `ObjectNameInformation = 1` (query handle name)
  - `ObjectTypeInformation = 2` (query handle type)
  - `StatusInfoLengthMismatch = 0xC0000004` (buffer too small - retry with larger buffer)
  - `DuplicateCloseSource = 0x00000001` (key flag to close remote handles)

#### enumerator.go - Handle Discovery
- **EnumerateProcessHandles()**: Main entry point
  - Queries `SystemExtendedHandleInformation` with dynamic buffer sizing (starts at 1MB)
  - Parses unsafe pointer arithmetic to extract `SystemHandleTableEntryInfoEx` array
  - Filters handles by target PID and type (only Event type = 19)
  - For each handle, calls `NtQueryObject()` twice: first for name, then for type string
  - **Key Pattern**: Unsafe pointer arithmetic at fixed offsets due to struct layout
  - **Error Handling**: Gracefully skips handles that fail name/type queries (some are inaccessible)

#### closer.go - Handle Closing
- **closeRemoteHandle()**: Uses `NtDuplicateObject()` with `DuplicateCloseSource` flag
  - This is the key trick: set target process to NULL and use close source flag
  - No actual duplication occurs; handle is closed in source process
- **CloseHandlesByName()**: Finds handles by name, closes each, returns count
- **findHandlesByName()**: Wrapper combining enumeration and filtering logic

### 3. D2R Constants (`pkg/d2r/`)
- **ProcessName**: `"D2R.exe"`
- **AgentProcessName**: `"Agent.exe"`  
- **SingleInstanceEventName**: `"DiabloII Check For Other Instances"`
- **Stored in separate package for reusability**

### 4. CLI & Logging (`cmd/multiablo/main.go`)
- **Cobra framework** for command structure (future extensibility)
- **Zap logging library** with dual modes:
  - Verbose: colorized development output with level names
  - Production: minimal console output without timestamps/levels
- Two main workflows:
  1. **Handle closing loop** (`handleCloserLoop`): Periodic checks (1s interval) to close single-instance handles
  2. **Agent killer loop** (`agentKillerLoop`): Periodic checks (10s interval) with intelligent uptime-based termination
     - Uses `checkAndKillAgentProcess()` helper function
     - Only kills Agent.exe processes with uptime ≥ 7 seconds
     - Maximizes Battle.net launcher availability for "Start Game" action
     - Executes immediately on startup, then periodically
- **Pattern**: Both use ticker with goroutine + select + chan coordination

---

## Critical Developer Workflows

### Building
```bash
# Standard Go build (outputs multiablo.exe in current directory)
go build -o multiablo.exe ./cmd/multiablo

# For building with Windows resources/manifest/version info:
# First install go-winres: go install github.com/tc-hib/go-winres@latest
# Then run: go-winres simply --arch amd64 --in cmd/multiablo/winres/winres.json --out cmd/multiablo/
```

### Testing
```bash
# Run test suite (test/test.go is a manual integration test, not automated)
go test ./...

# Manual integration testing requires:
# 1. Launch D2R.exe from Battle.net
# 2. Run: go run ./test/test.go
# This spawns background goroutines for continuous monitoring
```

### Running
```bash
# Must run as Administrator (Windows API requirements)
# Basic: finds and closes single-instance handle for all D2R processes
multiablo.exe

# Verbose mode: shows detailed debug logging
multiablo.exe -v
```

---

## Project-Specific Conventions & Patterns

### 1. Windows API Interaction
- **Unsafe pointer arithmetic throughout**: structs are unsafe pointers cast from raw buffer data
  - Example: `handleInfo := (*SystemExtendedHandleInformationEx)(unsafe.Pointer(&buffer[0]))`
  - Pattern: Calculate offsets with `unsafe.Sizeof()` for array traversal
- **Dynamic buffer allocation**: Many Windows Info queries use variable-length structures
  - Start with 1MB, retry with larger size if `StatusInfoLengthMismatch` returned
  - Store `returnLength` for proper sizing

### 2. Handle Enumeration Loop Pattern
```go
// Standard pattern across enumerator.go:
for i := 0; i < numberOfHandles; i++ {
    handlePtr := uintptr(unsafe.Pointer(&handleInfo.Handles[0])) + uintptr(i) * handleEntrySize
    entry := (*SystemHandleTableEntryInfoEx)(unsafe.Pointer(handlePtr))
    // Process entry...
}
```

### 3. Error Handling Philosophy
- **Graceful degradation**: If a single handle query fails, skip and continue
  - Some system processes' handles are inaccessible even with admin rights
  - Log the error but don't crash the entire enumeration
- **Return (count, error)** pattern: Return how many succeeded even if some failed

### 4. Logging Patterns
- Use `zap.String()`, `zap.Uint32()`, `zap.Int()` for structured fields
- No formatted strings in log calls; use `zap` field functions
- Example: `logger.Info("Found handles", zap.Int("count", len(handles)))`

### 5. Goroutine Coordination
- Use `chan struct{}` for stop signals (empty struct = no data, just signaling)
- Pattern: ticker + select with stop channel
```go
select {
case <-stopChan:
    return
case <-ticker.C:
    // Do work
}
```

---

## Integration Points & External Dependencies

### Windows API Dependencies
- **golang.org/x/sys/windows**: Official Windows API bindings (OpenProcess, CloseHandle, etc.)
- **syscall package**: Direct syscall invocations for undocumented ntdll functions
- **Undocumented APIs**: `NtQuerySystemInformation`, `NtQueryObject`, `NtDuplicateObject`
  - These are not officially documented by Microsoft but required for handle manipulation
  - Dependency on ntdll.dll (always present on Windows)

### Go Module Dependencies
- **github.com/spf13/cobra** (v1.10.2): CLI framework
- **go.uber.org/zap** (v1.27.1): Structured logging
- **golang.org/x/sys** (v0.39.0): Windows API bindings

### Cross-Component Communication
- **One-way flow**: Process → Handle Discovery → Handle Closing → Logging
- No circular dependencies or shared mutable state
- Processes are discovered fresh each iteration
- Handles are enumerated from current process snapshot

---

## Key Files to Understand First

1. **[internal/handle/winapi.go](internal/handle/winapi.go)** - Windows API definitions and constants
2. **[internal/handle/enumerator.go](internal/handle/enumerator.go)** - Core handle discovery logic (293 lines, most complex)
3. **[cmd/multiablo/main.go](cmd/multiablo/main.go)** - CLI entry point and main loops
4. **[internal/process/finder.go](internal/process/finder.go)** - Process enumeration and uptime tracking
5. **[pkg/d2r/constants.go](pkg/d2r/constants.go)** - Handle name and process name constants

---

## Common Pitfalls to Avoid

1. **Buffer size mismanagement**: Always check for `StatusInfoLengthMismatch` and retry with larger buffer
2. **Pointer arithmetic errors**: Ensure `handleEntrySize` is correct before calculating offsets
3. **Handle type filtering**: Event handle type index is 19 on most systems but OS-dependent; test before hardcoding
4. **Admin rights**: All handle operations require Administrator privileges; program must run elevated
5. **Process architecture mismatch**: 64-bit program can only safely manipulate 64-bit process handles (and vice versa)
6. **Goroutine leaks**: Always defer ticker.Stop() and ensure stop channel is closed

---

## Feature Roadmap Context

**Phase 1 (Current)**: Handle closing automation ✅
**Phase 2 (Planned)**: Auto-launcher with configurable instance count
**Phase 3 (Future)**: GUI, crash recovery, process monitoring

When adding new features, maintain the layered architecture: processes → handles → actions → logging.
