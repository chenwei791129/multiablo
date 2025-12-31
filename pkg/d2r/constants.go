// Package d2r provides constants and definitions specific to Diablo II: Resurrected.
package d2r

const (
	// ProcessName is the D2R executable name
	ProcessName = "D2R.exe"

	// AgentProcessName is the Battle.net Update Agent executable name
	AgentProcessName = "Agent.exe"

	// DefaultAgentPath is the default installation path of Agent.exe
	// This is used as a fallback when the running process path cannot be determined
	// The actual path is retrieved dynamically from the running process when available
	DefaultAgentPath = `C:\ProgramData\Battle.net\Agent\Agent.exe`

	// SingleInstanceEventName is the event handle name used by D2R to prevent multiple instances
	// Note: The actual handle name includes a session prefix like "\Sessions\1\BaseNamedObjects\"
	SingleInstanceEventName = "DiabloII Check For Other Instances"
)
