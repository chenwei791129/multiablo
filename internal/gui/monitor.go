package gui

import (
	"fmt"
	"sync"
	"time"

	"github.com/chenwei791129/multiablo/internal/handle"
	"github.com/chenwei791129/multiablo/internal/process"
	"github.com/chenwei791129/multiablo/pkg/d2r"
)

// ProcessInfo holds information about a monitored process
type ProcessInfo struct {
	PID          uint32
	Uptime       time.Duration
	HandleClosed bool
}

// MonitorStatus holds the current monitoring status
type MonitorStatus struct {
	D2RProcesses   []ProcessInfo
	AgentProcesses []ProcessInfo
	HandlesClosed  int
	AgentsKilled   int
	Event          string
}

// Monitor handles the background monitoring logic
type Monitor struct {
	stopChan chan struct{}
	statusCh chan MonitorStatus
	window   *MainWindow

	// Statistics
	totalHandlesClosed int
	totalAgentsKilled  int
	mu                 sync.Mutex

	// Running state
	running bool
	wg      sync.WaitGroup
}

// NewMonitor creates a new monitor instance
func NewMonitor(window *MainWindow) *Monitor {
	return &Monitor{
		window:   window,
		statusCh: make(chan MonitorStatus, 10),
	}
}

// Start begins the monitoring loops
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	m.wg.Add(3)
	go func() {
		defer m.wg.Done()
		m.handleCloserLoop()
	}()
	go func() {
		defer m.wg.Done()
		m.agentKillerLoop()
	}()
	go func() {
		defer m.wg.Done()
		m.statusUpdateLoop()
	}()
}

// Stop stops the monitoring loops
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	close(m.stopChan)
	m.mu.Unlock()

	// Wait for all goroutines to finish
	m.wg.Wait()
}

// IsRunning returns whether the monitor is running
func (m *Monitor) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// handleCloserLoop continuously monitors D2R processes and closes their single-instance handles
func (m *Monitor) handleCloserLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkD2RProcesses()
		}
	}
}

// checkD2RProcesses finds D2R processes and closes their single-instance handles
func (m *Monitor) checkD2RProcesses() {
	processes, err := process.FindProcessesByName(d2r.ProcessName)
	if err != nil {
		return
	}

	var d2rInfos []ProcessInfo
	for _, proc := range processes {
		info := ProcessInfo{
			PID:          proc.PID,
			HandleClosed: false,
		}

		// Try to close handles
		closedCount, err := handle.CloseHandlesByName(proc.PID, d2r.SingleInstanceEventName)
		if err == nil && closedCount > 0 {
			info.HandleClosed = true
			m.mu.Lock()
			m.totalHandlesClosed += closedCount
			m.mu.Unlock()

			m.sendStatus(MonitorStatus{
				Event: fmt.Sprintf("Closed %d handle(s) for D2R.exe (PID: %d)", closedCount, proc.PID),
			})
		}

		d2rInfos = append(d2rInfos, info)
	}

	// Send status update
	m.mu.Lock()
	totalClosed := m.totalHandlesClosed
	m.mu.Unlock()

	m.sendStatus(MonitorStatus{
		D2RProcesses:  d2rInfos,
		HandlesClosed: totalClosed,
	})
}

// agentKillerLoop continuously monitors and kills Agent.exe processes
func (m *Monitor) agentKillerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkAgentProcesses()
		}
	}
}

// checkAgentProcesses finds Agent.exe processes and kills them if needed
func (m *Monitor) checkAgentProcesses() {
	processes, err := process.FindProcessesByName(d2r.AgentProcessName)
	if err != nil {
		return
	}

	var agentInfos []ProcessInfo
	for _, proc := range processes {
		uptime, _ := process.GetProcessUptime(proc.PID)
		info := ProcessInfo{
			PID:    proc.PID,
			Uptime: uptime,
		}
		agentInfos = append(agentInfos, info)
	}

	// Check if we should kill Agent.exe
	if len(processes) > 0 {
		uptime, _ := process.GetProcessOldestUptimeByName(d2r.AgentProcessName)
		if uptime >= (time.Second * 7) {
			// Get Agent.exe path before killing
			agentPath := ""
			path, err := process.GetProcessExecutablePath(processes[0].PID)
			if err == nil {
				agentPath = path
			}
			if agentPath == "" {
				agentPath = d2r.DefaultAgentPath
			}

			// Kill Agent.exe processes
			killedCount, err := process.KillProcessesByName(d2r.AgentProcessName)
			if err == nil && killedCount > 0 {
				m.mu.Lock()
				m.totalAgentsKilled += killedCount
				m.mu.Unlock()

				m.sendStatus(MonitorStatus{
					Event: fmt.Sprintf("Terminated %d Agent.exe process(es)", killedCount),
				})

				// Relaunch Agent.exe
				err = process.LaunchProcess(agentPath)
				if err != nil {
					m.sendStatus(MonitorStatus{
						Event: fmt.Sprintf("Failed to relaunch Agent.exe: %v", err),
					})
				} else {
					m.sendStatus(MonitorStatus{
						Event: "Relaunched Agent.exe successfully",
					})
				}
			}
		}
	}

	// Send status update
	m.mu.Lock()
	totalKilled := m.totalAgentsKilled
	m.mu.Unlock()

	m.sendStatus(MonitorStatus{
		AgentProcesses: agentInfos,
		AgentsKilled:   totalKilled,
	})
}

// sendStatus sends a status update to the channel
func (m *Monitor) sendStatus(status MonitorStatus) {
	select {
	case m.statusCh <- status:
	default:
		// Channel full, skip this update
	}
}

// statusUpdateLoop processes status updates and updates the UI
func (m *Monitor) statusUpdateLoop() {
	var lastD2RProcesses []ProcessInfo
	var lastAgentProcesses []ProcessInfo
	var lastHandlesClosed int
	var lastAgentsKilled int

	// Use a ticker to throttle UI updates
	updateTicker := time.NewTicker(500 * time.Millisecond)
	defer updateTicker.Stop()

	needsUpdate := false

	for {
		select {
		case <-m.stopChan:
			return
		case status := <-m.statusCh:
			// Update D2R status if changed
			if status.D2RProcesses != nil {
				lastD2RProcesses = status.D2RProcesses
				needsUpdate = true
			}
			if status.HandlesClosed > 0 {
				lastHandlesClosed = status.HandlesClosed
				needsUpdate = true
			}

			// Update Agent status if changed
			if status.AgentProcesses != nil {
				lastAgentProcesses = status.AgentProcesses
				needsUpdate = true
			}
			if status.AgentsKilled > 0 {
				lastAgentsKilled = status.AgentsKilled
				needsUpdate = true
			}

			// Log events immediately
			if status.Event != "" {
				m.window.AppendLog(status.Event)
			}
		case <-updateTicker.C:
			// Throttled UI update
			if needsUpdate {
				m.updateD2RUI(lastD2RProcesses, lastHandlesClosed)
				m.updateAgentUI(lastAgentProcesses, lastAgentsKilled)
				needsUpdate = false
			}
		}
	}
}

// updateD2RUI updates the D2R section of the UI
func (m *Monitor) updateD2RUI(processes []ProcessInfo, handlesClosed int) {
	processList := ""
	for _, p := range processes {
		status := "monitoring"
		if p.HandleClosed {
			status = "handle closed"
		}
		if processList != "" {
			processList += "\n"
		}
		processList += fmt.Sprintf("PID %d - %s", p.PID, status)
	}

	m.window.UpdateD2RStatus(len(processes), processList, handlesClosed)
}

// updateAgentUI updates the Agent section of the UI
func (m *Monitor) updateAgentUI(processes []ProcessInfo, agentsKilled int) {
	processList := ""
	for _, p := range processes {
		if processList != "" {
			processList += "\n"
		}
		processList += fmt.Sprintf("PID %d - uptime: %.1fs", p.PID, p.Uptime.Seconds())
	}

	m.window.UpdateAgentStatus(len(processes), processList, agentsKilled)
}
