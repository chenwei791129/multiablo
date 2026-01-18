package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	windowWidth  = 650
	windowHeight = 550
	maxLogLines  = 500
)

// MainWindow represents the main application window
type MainWindow struct {
	window fyne.Window

	// UI Components - D2R monitoring
	d2rCountLabel         *widget.Label
	d2rProcessList        *widget.Label
	d2rHandlesClosedLabel *widget.Label

	// UI Components - Agent monitoring
	agentCountLabel  *widget.Label
	agentProcessList *widget.Label
	agentKilledLabel *widget.Label

	// UI Components - Log
	logLabel *widget.Label

	// UI Components - Controls
	startStopBtn *widget.Button
	clearLogBtn  *widget.Button

	// Data Binding
	d2rCountBinding     binding.String
	d2rProcessBinding   binding.String
	d2rHandlesBinding   binding.String
	agentCountBinding   binding.String
	agentProcessBinding binding.String
	agentKilledBinding  binding.String
	logBinding          binding.String

	// State
	isMonitoring bool
	logLines     []string

	// Monitor
	monitor *Monitor

	// Synchronization
	mu sync.Mutex
}

// NewMainWindow creates and configures the main window
func NewMainWindow(app fyne.App) *MainWindow {
	w := &MainWindow{
		isMonitoring: false,
		logLines:     make([]string, 0, maxLogLines),
	}
	w.window = app.NewWindow(AppTitle)
	w.window.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.window.CenterOnScreen()

	// Initialize bindings
	w.d2rCountBinding = binding.NewString()
	w.d2rProcessBinding = binding.NewString()
	w.d2rHandlesBinding = binding.NewString()
	w.agentCountBinding = binding.NewString()
	w.agentProcessBinding = binding.NewString()
	w.agentKilledBinding = binding.NewString()
	w.logBinding = binding.NewString()

	// Handle window close
	w.window.SetCloseIntercept(func() {
		w.mu.Lock()
		monitoring := w.isMonitoring
		w.mu.Unlock()

		if w.monitor != nil && monitoring {
			w.monitor.Stop()
		}
		w.window.Close()
	})

	w.createUI()
	return w
}

// createUI builds the user interface
func (w *MainWindow) createUI() {
	// D2R.exe monitoring section
	w.d2rCountBinding.Set("Detected processes: 0")
	w.d2rCountLabel = widget.NewLabelWithData(w.d2rCountBinding)

	w.d2rProcessBinding.Set("No D2R.exe processes detected")
	w.d2rProcessList = widget.NewLabelWithData(w.d2rProcessBinding)
	w.d2rProcessList.Wrapping = fyne.TextWrapWord

	w.d2rHandlesBinding.Set("Total handles closed: 0")
	w.d2rHandlesClosedLabel = widget.NewLabelWithData(w.d2rHandlesBinding)

	d2rCard := widget.NewCard("D2R.exe Monitor", "",
		container.NewVBox(
			w.d2rCountLabel,
			w.d2rProcessList,
			w.d2rHandlesClosedLabel,
		),
	)

	// Agent.exe monitoring section
	w.agentCountBinding.Set("Detected processes: 0")
	w.agentCountLabel = widget.NewLabelWithData(w.agentCountBinding)

	w.agentProcessBinding.Set("No Agent.exe processes detected")
	w.agentProcessList = widget.NewLabelWithData(w.agentProcessBinding)
	w.agentProcessList.Wrapping = fyne.TextWrapWord

	w.agentKilledBinding.Set("Total processes terminated: 0")
	w.agentKilledLabel = widget.NewLabelWithData(w.agentKilledBinding)

	agentCard := widget.NewCard("Agent.exe Monitor", "",
		container.NewVBox(
			w.agentCountLabel,
			w.agentProcessList,
			w.agentKilledLabel,
		),
	)

	// Log section with scroll
	w.logLabel = widget.NewLabelWithData(w.logBinding)
	w.logLabel.Wrapping = fyne.TextWrapWord
	// Use TextStyle to simulate monospace if desired, or keep default
	// w.logLabel.TextStyle = fyne.TextStyle{Monospace: true}

	logScroll := container.NewScroll(w.logLabel)
	logScroll.SetMinSize(fyne.NewSize(600, 200))

	logCard := widget.NewCard("Activity Log", "", logScroll)

	// Control buttons
	w.startStopBtn = widget.NewButton("Start Monitoring", func() {
		w.onStartStopClick()
	})
	w.startStopBtn.Importance = widget.HighImportance

	w.clearLogBtn = widget.NewButton("Clear Log", func() {
		w.onClearLogClick()
	})

	controlBox := container.NewHBox(
		layout.NewSpacer(),
		w.startStopBtn,
		w.clearLogBtn,
		layout.NewSpacer(),
	)

	// Main layout with padding
	content := container.NewVBox(
		d2rCard,
		agentCard,
		logCard,
		widget.NewSeparator(),
		controlBox,
	)

	padded := container.NewPadded(content)
	w.window.SetContent(padded)

	// Initial log message
	w.appendLog("Multiablo GUI started")
	w.appendLog("Monitoring will start automatically...")
}

// Show displays the window
func (w *MainWindow) Show() {
	w.window.Show()
}

// StartMonitoringAutomatically starts monitoring when the app launches
func (w *MainWindow) StartMonitoringAutomatically() {
	// Execute immediately on the main thread (during startup phase)
	// Removing goroutine and sleep prevents data races with UI initialization
	w.onStartStopClick()
}

// onStartStopClick handles the start/stop button click
func (w *MainWindow) onStartStopClick() {
	w.mu.Lock()
	if !w.isMonitoring {
		w.isMonitoring = true
		w.mu.Unlock()

		w.startStopBtn.SetText("Stop Monitoring")
		w.startStopBtn.Importance = widget.DangerImportance
		w.appendLog("Monitoring started...")

		// Create and start monitor
		if w.monitor == nil {
			w.monitor = NewMonitor(w)
		}
		w.monitor.Start()
	} else {
		w.isMonitoring = false
		w.mu.Unlock()

		w.startStopBtn.SetText("Start Monitoring")
		w.startStopBtn.Importance = widget.HighImportance
		w.appendLog("Monitoring stopped.")

		// Stop monitor
		if w.monitor != nil {
			w.monitor.Stop()
		}
	}
}

// onClearLogClick handles the clear log button click
func (w *MainWindow) onClearLogClick() {
	w.mu.Lock()
	w.logLines = w.logLines[:0]
	w.logBinding.Set("")
	w.mu.Unlock()
}

// appendLog adds a timestamped message to the log
func (w *MainWindow) appendLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	w.mu.Lock()
	// Add new line to slice
	w.logLines = append(w.logLines, logLine)

	// Trim log if it exceeds max lines
	if len(w.logLines) > maxLogLines {
		w.trimLogLocked()
	}

	// Update Binding
	fullLog := strings.Join(w.logLines, "\n")
	w.logBinding.Set(fullLog)
	w.mu.Unlock()
}

// trimLog removes older log entries to keep the log size manageable
func (w *MainWindow) trimLog() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.trimLogLocked()
}

// trimLogLocked removes older log entries (caller must hold w.mu)
func (w *MainWindow) trimLogLocked() {
	keepLines := maxLogLines / 2
	if len(w.logLines) > keepLines {
		w.logLines = w.logLines[len(w.logLines)-keepLines:]
	}
}

// UpdateD2RStatus updates the D2R monitoring display
func (w *MainWindow) UpdateD2RStatus(processCount int, processList string, handlesClosed int) {
	w.d2rCountBinding.Set(fmt.Sprintf("Detected processes: %d", processCount))
	if processList == "" {
		w.d2rProcessBinding.Set("No D2R.exe processes detected")
	} else {
		w.d2rProcessBinding.Set(processList)
	}
	w.d2rHandlesBinding.Set(fmt.Sprintf("Total handles closed: %d", handlesClosed))
}

// UpdateAgentStatus updates the Agent monitoring display
func (w *MainWindow) UpdateAgentStatus(processCount int, processList string, agentsKilled int) {
	w.agentCountBinding.Set(fmt.Sprintf("Detected processes: %d", processCount))
	if processList == "" {
		w.agentProcessBinding.Set("No Agent.exe processes detected")
	} else {
		w.agentProcessBinding.Set(processList)
	}
	w.agentKilledBinding.Set(fmt.Sprintf("Total processes terminated: %d", agentsKilled))
}

// IsMonitoring returns the current monitoring state
func (w *MainWindow) IsMonitoring() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isMonitoring
}

// AppendLog exposes the log append function for external use
func (w *MainWindow) AppendLog(message string) {
	w.appendLog(message)
}
