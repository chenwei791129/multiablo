package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	logText *widget.RichText

	// UI Components - Controls
	startStopBtn *widget.Button
	clearLogBtn  *widget.Button

	// State
	isMonitoring bool
	logLines     []string

	// Monitor
	monitor *Monitor
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

	// Handle window close
	w.window.SetCloseIntercept(func() {
		if w.monitor != nil && w.isMonitoring {
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
	w.d2rCountLabel = widget.NewLabel("Detected processes: 0")
	w.d2rProcessList = widget.NewLabel("No D2R.exe processes detected")
	w.d2rProcessList.Wrapping = fyne.TextWrapWord
	w.d2rHandlesClosedLabel = widget.NewLabel("Total handles closed: 0")

	d2rCard := widget.NewCard("D2R.exe Monitor", "",
		container.NewVBox(
			w.d2rCountLabel,
			w.d2rProcessList,
			w.d2rHandlesClosedLabel,
		),
	)

	// Agent.exe monitoring section
	w.agentCountLabel = widget.NewLabel("Detected processes: 0")
	w.agentProcessList = widget.NewLabel("No Agent.exe processes detected")
	w.agentProcessList.Wrapping = fyne.TextWrapWord
	w.agentKilledLabel = widget.NewLabel("Total processes terminated: 0")

	agentCard := widget.NewCard("Agent.exe Monitor", "",
		container.NewVBox(
			w.agentCountLabel,
			w.agentProcessList,
			w.agentKilledLabel,
		),
	)

	// Log section with scroll
	w.logText = widget.NewRichText()
	w.logText.Wrapping = fyne.TextWrapWord

	logScroll := container.NewScroll(w.logText)
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
	w.appendLog("Click 'Start Monitoring' to begin")
}

// Show displays the window
func (w *MainWindow) Show() {
	w.window.Show()
}

// StartMonitoringAutomatically starts monitoring when the app launches
func (w *MainWindow) StartMonitoringAutomatically() {
	// Auto-start monitoring after a brief delay to let UI render
	go func() {
		time.Sleep(100 * time.Millisecond)
		w.onStartStopClick()
	}()
}

// onStartStopClick handles the start/stop button click
func (w *MainWindow) onStartStopClick() {
	if !w.isMonitoring {
		w.isMonitoring = true
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
	w.logLines = w.logLines[:0]
	w.logText.Segments = nil
	w.logText.Refresh()
}

// appendLog adds a timestamped message to the log
func (w *MainWindow) appendLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	// Add new line to slice
	w.logLines = append(w.logLines, logLine)

	// Trim log if it exceeds max lines
	if len(w.logLines) > maxLogLines {
		w.trimLog()
	}

	// Update RichText with single TextSegment containing all lines
	w.logText.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Text: strings.Join(w.logLines, "\n"),
		},
	}
	w.logText.Refresh()
}

// trimLog removes older log entries to keep the log size manageable
func (w *MainWindow) trimLog() {
	keepLines := maxLogLines / 2
	if len(w.logLines) > keepLines {
		w.logLines = w.logLines[len(w.logLines)-keepLines:]
	}
}

// UpdateD2RStatus updates the D2R monitoring display
func (w *MainWindow) UpdateD2RStatus(processCount int, processList string, handlesClosed int) {
	w.d2rCountLabel.SetText(fmt.Sprintf("Detected processes: %d", processCount))
	if processList == "" {
		w.d2rProcessList.SetText("No D2R.exe processes detected")
	} else {
		w.d2rProcessList.SetText(processList)
	}
	w.d2rHandlesClosedLabel.SetText(fmt.Sprintf("Total handles closed: %d", handlesClosed))
}

// UpdateAgentStatus updates the Agent monitoring display
func (w *MainWindow) UpdateAgentStatus(processCount int, processList string, agentsKilled int) {
	w.agentCountLabel.SetText(fmt.Sprintf("Detected processes: %d", processCount))
	if processList == "" {
		w.agentProcessList.SetText("No Agent.exe processes detected")
	} else {
		w.agentProcessList.SetText(processList)
	}
	w.agentKilledLabel.SetText(fmt.Sprintf("Total processes terminated: %d", agentsKilled))
}

// IsMonitoring returns the current monitoring state
func (w *MainWindow) IsMonitoring() bool {
	return w.isMonitoring
}

// AppendLog exposes the log append function for external use
func (w *MainWindow) AppendLog(message string) {
	w.appendLog(message)
}
