# Multiablo Copilot Instructions

## Project Overview

**Multiablo** is a Windows-based D2R (Diablo II: Resurrected) multi-instance launcher written in Go with a Fyne GUI. It automatically closes the single-instance Event Handle (`DiabloII Check For Other Instances`) that prevents D2R from running multiple instances, enabling users to launch multiple D2R clients simultaneously.

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
  - `GetProcessExecutablePath()` - Get the full path of a process executable
  - `LaunchProcess()` - Start a new process by path
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
- **DefaultAgentPath**: Default path to Agent.exe for relaunching
- **Stored in separate package for reusability**

### 4. GUI Layer (`internal/gui/`)
The application uses Fyne v2 for the graphical user interface.

#### app.go - Application Entry Point
- **App struct**: Wraps the Fyne application
- **NewApp()**: Creates a new GUI application with unique AppID
- **Run()**: Starts the application and auto-starts monitoring
- **AppID**: `"com.github.chenwei791129.multiablo"`
- **AppTitle**: `"Multiablo - D2R Multi-Instance Helper"`

#### mainwindow.go - Main Window UI
- **MainWindow struct**: Contains all UI components and state
- **Key UI Components**:
  - D2R.exe monitoring card: process count, process list, handles closed count
  - Agent.exe monitoring card: process count, process list (with uptime), agents killed count
  - Activity log (scrollable multi-line entry, max 500 lines with auto-trim)
  - Start/Stop monitoring button
  - Clear log button
- **Key Functions**:
  - `NewMainWindow()` - Creates and configures the main window
  - `createUI()` - Builds the user interface layout
  - `StartMonitoringAutomatically()` - Auto-starts monitoring on app launch
  - `UpdateD2RStatus()` - Updates D2R monitoring display
  - `UpdateAgentStatus()` - Updates Agent monitoring display
  - `AppendLog()` - Adds timestamped messages to the activity log

#### monitor.go - Background Monitoring Logic
- **Monitor struct**: Manages background monitoring goroutines
- **MonitorStatus struct**: Holds current monitoring status for UI updates
- **ProcessInfo struct**: Holds PID, uptime, and handle status for processes
- **Two monitoring loops** (both use 1s ticker):
  1. **handleCloserLoop**: Monitors D2R.exe processes and closes single-instance handles
  2. **agentKillerLoop**: Monitors Agent.exe processes and terminates them after 7 seconds uptime, then relaunches Agent.exe
- **statusUpdateLoop**: Processes status updates and throttles UI updates (500ms)
- **Pattern**: Uses channels for communication and mutex for thread-safe counters

### 5. Entry Point (`cmd/multiablo/main.go`)
- Simple entry point that creates and runs the GUI application
- No CLI flags - pure GUI application
- Build with `-H windowsgui` flag to hide console window

---

## Critical Developer Workflows

### Building
```bash
# Install dependencies
go mod tidy

# Generate Windows resources (.syso file)
cd cmd/multiablo && go tool go-winres make --arch amd64 && cd ../..

# Build the application (requires MinGW-w64 for CGO/Fyne)
CGO_ENABLED=1 go build -ldflags="-s -w -H windowsgui" -o multiablo.exe ./cmd/multiablo
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
# Simply run the executable - GUI will appear and auto-start monitoring
multiablo.exe
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

### 4. GUI Update Patterns
- Use channels for communication between monitor goroutines and UI
- Throttle UI updates (500ms) to prevent flickering
- Log events immediately but batch status updates
- Use `sync.Mutex` for thread-safe access to counters

### 5. Goroutine Coordination
- Use `chan struct{}` for stop signals (empty struct = no data, just signaling)
- Pattern: ticker + select with stop channel
```go
select {
case <-m.stopChan:
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
- **fyne.io/fyne/v2**: Cross-platform GUI toolkit (requires CGO and MinGW-w64 on Windows)
- **golang.org/x/sys** (v0.39.0): Windows API bindings

### Cross-Component Communication
- **GUI → Monitor**: Start/Stop commands, window reference for updates
- **Monitor → GUI**: Status updates via channel, log messages
- **One-way flow**: Process → Handle Discovery → Handle Closing → UI Update
- No circular dependencies or shared mutable state (except counters with mutex)

---

## Key Files to Understand First

1. **[internal/gui/app.go](../internal/gui/app.go)** - Application entry point and lifecycle
2. **[internal/gui/mainwindow.go](../internal/gui/mainwindow.go)** - Main window UI and layout
3. **[internal/gui/monitor.go](../internal/gui/monitor.go)** - Background monitoring logic
4. **[internal/handle/winapi.go](../internal/handle/winapi.go)** - Windows API definitions and constants
5. **[internal/handle/enumerator.go](../internal/handle/enumerator.go)** - Core handle discovery logic
6. **[internal/process/finder.go](../internal/process/finder.go)** - Process enumeration and uptime tracking
7. **[pkg/d2r/constants.go](../pkg/d2r/constants.go)** - Handle name and process name constants

---

## Common Pitfalls to Avoid

1. **Buffer size mismanagement**: Always check for `StatusInfoLengthMismatch` and retry with larger buffer
2. **Pointer arithmetic errors**: Ensure `handleEntrySize` is correct before calculating offsets
3. **Handle type filtering**: Event handle type index is 19 on most systems but OS-dependent; test before hardcoding
4. **Admin rights**: All handle operations require Administrator privileges; program must run elevated
5. **Process architecture mismatch**: 64-bit program can only safely manipulate 64-bit process handles (and vice versa)
6. **Goroutine leaks**: Always defer ticker.Stop() and ensure stop channel is closed
7. **CGO/Fyne build issues**: Ensure MinGW-w64 is installed and CGO_ENABLED=1 is set
8. **UI thread safety**: Always use channels for goroutine-to-UI communication; never update UI directly from background goroutines

---

## Feature Roadmap Context

**Phase 1**: Handle closing automation ✅
**Phase 2**: GUI interface with Fyne ✅ (Current)
**Phase 3 (Future)**: Auto-launcher, configuration, process monitoring enhancements

When adding new features, maintain the layered architecture: processes → handles → monitor → UI.

---

## Important Technical Decisions

### Why Go?
- Excellent Windows API bindings via `golang.org/x/sys/windows`
- Single executable, no runtime dependencies (except MinGW runtime for Fyne)
- High performance with low memory footprint
- Cross-compilation support for different architectures

### Why Fyne for GUI?
- Cross-platform (though this app is Windows-only)
- Modern, material-design-inspired look
- Pure Go API (easier to integrate than CGo-heavy alternatives)
- Active development and community support
- Easy to create cards, buttons, and scrollable log areas

### Why Undocumented Windows APIs?
- `NtQuerySystemInformation` and `NtQueryObject` are the only way to enumerate system handles
- Tools like Process Explorer use these same APIs
- Risk: Windows updates could change behavior (but historically very stable)
- These are from `ntdll.dll` which is always present on Windows

### Why Uptime-Based Agent.exe Termination?
- Allows Battle.net launcher enough time to execute "Start Game" action
- 7-second threshold found through testing to be optimal balance
- Maximizes launcher availability while preventing interference
- Relaunches Agent.exe after termination to prevent Battle.net lockup

---

## Known Limitations and Challenges

### Platform Limitations
1. **Windows Only**: D2R itself is Windows-only, so this is an inherent constraint
2. **Administrator Rights Required**: Handle manipulation requires elevated privileges
3. **Architecture Matching**: 64-bit program can only manipulate 64-bit process handles safely
4. **MinGW-w64 Required**: Fyne GUI requires CGO which needs MinGW-w64 toolchain

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
# Install MinGW-w64 (required for Fyne/CGO)
# On Windows, use MSYS2 or download from mingw-w64.org

# Install go-winres
go install github.com/tc-hib/go-winres@latest
```

**2. Generate Windows Resources**:
```bash
# Generate .syso files with manifest and version info
cd cmd/multiablo && go tool go-winres make --arch amd64 && cd ../..
```
This creates `.syso` file in same directory as `main.go`

**3. Build Executable**:
```bash
# Build with GUI (no console window)
CGO_ENABLED=1 go build -ldflags="-s -w -H windowsgui" -o multiablo.exe ./cmd/multiablo
```

**Important Notes**:
- `.syso` files must be in same directory as `main.go`
- After modifying `winres.json`, re-run `go-winres` to regenerate
- `.syso` files are in `.gitignore` (generated locally)
- Go build automatically links `.syso` files during compilation
- Result: Windows shows UAC prompt when user launches the program
- `-H windowsgui` flag hides the console window for a clean GUI experience

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

### 2025-12-20: UAC Integration
- Added Windows Manifest for auto-elevation
- Integrated `go-winres` build tooling
- Automated administrator privilege request

### 2025-12-28: Intelligent Agent.exe Termination
- Added process uptime tracking functions
- Implemented 7-second uptime threshold for Agent.exe killing
- Refactored into `checkAndKillAgentProcess()` helper
- Maximized Battle.net launcher availability

### 2026-01: GUI Implementation with Fyne
- Replaced CLI interface with Fyne GUI
- Added D2R.exe and Agent.exe monitoring cards
- Implemented scrollable activity log with auto-trim
- Added Start/Stop monitoring control
- Auto-start monitoring on application launch
- Added Agent.exe relaunch after termination

---

## Environment Requirements

### Development Environment
- Go 1.21 or higher
- Windows 10/11 (64-bit recommended)
- MinGW-w64 (for CGO/Fyne compilation)
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
- [fyne.io/fyne/v2](https://pkg.go.dev/fyne.io/fyne/v2) - Cross-platform GUI toolkit

---

## Development Notes and Tips

### Handle Enumeration Best Practices
- Always skip System process (PID 4) to avoid deadlocks
- Implement graceful error handling for inaccessible handles
- Consider timeout mechanisms for slow handle queries (future enhancement)
- Use dynamic buffer sizing for `NtQuerySystemInformation` results

### Testing Workflow
1. Build the application: `CGO_ENABLED=1 go build -ldflags="-s -w -H windowsgui" -o multiablo.exe ./cmd/multiablo`
2. Run `multiablo.exe` (UAC prompt will appear)
3. Verify GUI appears and monitoring auto-starts
4. Launch D2R from Battle.net launcher (first instance)
5. Check activity log for handle closing messages
6. Launch second D2R instance from Battle.net launcher
7. Confirm both instances run simultaneously
8. Check Agent.exe termination and relaunch behavior in activity log

---

## Project Status

**Current Version**: 2.0.0+
**Status**: ✅ GUI implementation complete
**License**: MIT License
**Repository**: https://github.com/chenwei791129/multiablo

### Completed Features
- ✅ Automatic D2R.exe process detection
- ✅ System-wide handle enumeration
- ✅ Single-instance handle closing
- ✅ Multiple D2R instance support
- ✅ UAC auto-elevation
- ✅ Intelligent Agent.exe termination with uptime checking
- ✅ Agent.exe relaunch after termination
- ✅ GUI interface with Fyne
- ✅ Real-time process monitoring display
- ✅ Activity log with auto-trim
- ✅ Start/Stop monitoring control
- ✅ Comprehensive documentation

### Future Enhancements (Consideration)
- [ ] Automatic D2R launch feature
- [ ] Configuration file support (YAML/JSON)
- [ ] System tray icon
- [ ] Process crash recovery
- [ ] Multi-window management (auto-arrange)
- [ ] GitHub Actions CI/CD for releases

---

**Last Updated**: 2026-01
**Maintained By**: chenwei791129
