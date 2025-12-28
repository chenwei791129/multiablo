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
# Standard Go build
go build -o multiablo.exe ./cmd/multiablo

# See "UAC and Build System Details" section for full build process with Windows resources
```

### Testing
```bash
# Run unit tests
go test ./...

# See "Development Notes and Tips > Testing Workflow" for detailed integration testing steps
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

---

## Important Technical Decisions

### Why Go?
- Excellent Windows API bindings via `golang.org/x/sys/windows`
- Single executable, no runtime dependencies
- High performance with low memory footprint
- Cross-compilation support for different architectures

### Why Undocumented Windows APIs?
- `NtQuerySystemInformation` and `NtQueryObject` are the only way to enumerate system handles
- Tools like Process Explorer use these same APIs
- Risk: Windows updates could change behavior (but historically very stable)
- These are from `ntdll.dll` which is always present on Windows

### Why Zap over Standard log?
- Zero-allocation performance
- Structured logging for better analysis
- Flexible configuration options
- Production-ready with minimal overhead

### Why Cobra?
- Industry-standard CLI framework in Go ecosystem
- Auto-generated help and usage information
- Easy to extend with subcommands
- Professional CLI experience out of the box

### Why Uptime-Based Agent.exe Termination?
- Allows Battle.net launcher enough time to execute "Start Game" action
- 7-second threshold found through testing to be optimal balance
- Maximizes launcher availability while preventing interference
- More intelligent than fixed-interval killing

---

## Known Limitations and Challenges

### Platform Limitations
1. **Windows Only**: D2R itself is Windows-only, so this is an inherent constraint
2. **Administrator Rights Required**: Handle manipulation requires elevated privileges
3. **Architecture Matching**: 64-bit program can only manipulate 64-bit process handles safely

### API and Compatibility Risks
1. **Handle Name Dependency**: If Blizzard changes the `DiabloII Check For Other Instances` handle name, program will fail
2. **Undocumented API Risks**: Future Windows updates could theoretically break `NtQuerySystemInformation` behavior
3. **Anti-Cheat Concerns**: Multi-boxing may violate game ToS; risk of account suspension (documented in README)

### Technical Challenges
1. **Handle Enumeration Deadlocks**: Querying certain handle types (especially named pipes) can cause deadlocks
   - Current mitigation: Skip System process (PID 4)
   - Skip handles that fail to open
   - **TODO**: Add timeout mechanism for handle queries
2. **Antivirus False Positives**: Handle manipulation tools are commonly flagged (expected behavior)

---

## UAC and Build System Details

### UAC Auto-Elevation
The program uses a Windows Manifest to automatically request administrator privileges:

**Configuration** (`cmd/multiablo/winres/winres.json`):
```json
{
  "RT_MANIFEST": {
    "execution-level": "requireAdministrator"
  }
}
```

### Complete Build Process

**1. Install Build Tools** (first time only):
```bash
go install github.com/tc-hib/go-winres@latest
```

**2. Generate Windows Resources**:
```bash
# Generate .syso files with manifest and version info
go-winres simply --arch amd64 --in cmd/multiablo/winres/winres.json --out cmd/multiablo/
```
This creates `rsrc_windows_amd64.syso` in same directory as `main.go`

**3. Build Executable**:
```bash
# Development build
go build -o multiablo.exe ./cmd/multiablo

# Release build (with optimization and version)
go build -ldflags="-s -w -X main.version=1.0.0" -o multiablo.exe ./cmd/multiablo
```

**Important Notes**:
- `.syso` files must be in same directory as `main.go`
- Filename format: `rsrc_windows_{arch}.syso`
- After modifying `winres.json`, re-run `go-winres` to regenerate
- `.syso` files are in `.gitignore` (generated locally)
- Go build automatically links `.syso` files during compilation
- Result: Windows shows UAC prompt when user launches the program

---

## Development History Summary

### 2025-12-20: Project Initialization
- Created GitHub repository
- Initialized Go module
- Planned implementation approach

### 2025-12-20: Core Feature Implementation (Milestone 1)
- Implemented Windows API wrappers
- Built handle enumeration and closing logic
- Created process discovery layer
- Developed main program logic with CLI

### 2025-12-20: Professional Refactoring
- Replaced `flag` with `cobra` for better CLI
- Replaced `fmt.Println` with `zap` structured logging
- Implemented dual logging modes (normal/verbose)
- Disabled stack traces for cleaner output

### 2025-12-20: Feature Simplification
- Removed `--close-only` flag
- Made program focus on handle closing only
- User launches D2R manually via Battle.net
- Program runs continuously in background

### 2025-12-20: UAC Integration
- Added Windows Manifest for auto-elevation
- Integrated `go-winres` build tooling
- Automated administrator privilege request

### 2025-12-28: Intelligent Agent.exe Termination
- Added process uptime tracking functions
- Implemented 7-second uptime threshold for Agent.exe killing
- Refactored into `checkAndKillAgentProcess()` helper
- Maximized Battle.net launcher availability

---

## Environment Requirements

### Development Environment
- Go 1.21 or higher
- Windows 10/11 (64-bit recommended)
- Git for version control
- `go-winres` tool for resource generation
- Administrator privileges (required for testing)

### Testing Requirements
- D2R installed and configured
- Battle.net launcher
- Multiple Battle.net accounts (optional, for multi-instance testing)

---

## Reference Materials

### Windows API Documentation
- [NtQuerySystemInformation](https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-ntquerysysteminformation) - System information queries
- [NtDuplicateObject](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-zwduplicateobject) - Handle duplication/closing
- [Process and Thread Functions](https://learn.microsoft.com/en-us/windows/win32/procthread/process-and-thread-functions) - Process management
- [GetProcessTimes](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getprocesstimes) - Process timing information

### Similar Projects (for reference)
- [Process Hacker](https://github.com/processhacker/processhacker) - Open-source process manager (handle manipulation reference)
- [handle.exe](https://learn.microsoft.com/en-us/sysinternals/downloads/handle) - Sysinternals handle viewer tool

### Go Library Documentation
- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) - Windows API bindings
- [go.uber.org/zap](https://pkg.go.dev/go.uber.org/zap) - High-performance logging
- [github.com/spf13/cobra](https://pkg.go.dev/github.com/spf13/cobra) - CLI framework

---

## Development Notes and Tips

### Handle Enumeration Best Practices
- Always skip System process (PID 4) to avoid deadlocks
- Implement graceful error handling for inaccessible handles
- Consider timeout mechanisms for slow handle queries (future enhancement)
- Use dynamic buffer sizing for `NtQuerySystemInformation` results

### Testing Workflow
1. Launch D2R from Battle.net launcher (first instance)
2. Run `multiablo.exe` with verbose flag: `multiablo.exe -v`
3. Verify handle closing messages appear in log
4. Launch second D2R instance from Battle.net launcher
5. Confirm both instances run simultaneously
6. Check Agent.exe termination behavior with uptime logging

---

## Project Status

**Current Version**: 1.0.0+
**Status**: ✅ Core features complete and tested
**License**: MIT License
**Repository**: https://github.com/chenwei791129/multiablo

### Completed Features
- ✅ Automatic D2R.exe process detection
- ✅ System-wide handle enumeration
- ✅ Single-instance handle closing
- ✅ Multiple D2R instance support
- ✅ UAC auto-elevation
- ✅ Structured logging with zap
- ✅ Professional CLI with cobra
- ✅ Intelligent Agent.exe termination with uptime checking
- ✅ Comprehensive documentation

### Future Enhancements (Consideration)
- [ ] Automatic D2R launch feature
- [ ] Configuration file support (YAML/JSON)
- [ ] GUI interface (fyne or walk)
- [ ] Process crash recovery
- [ ] Multi-window management (auto-arrange)
- [ ] GitHub Actions CI/CD for releases

---

**Last Updated**: 2025-12-28
**Maintained By**: chenwei791129
