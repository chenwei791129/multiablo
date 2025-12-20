package d2r

const (
	// ProcessName is the D2R executable name
	ProcessName = "D2R.exe"

	// AgentProcessName is the Battle.net Update Agent executable name
	AgentProcessName = "Agent.exe"

	// SingleInstanceEventName is the event handle name used by D2R to prevent multiple instances
	// Note: The actual handle name includes a session prefix like "\Sessions\1\BaseNamedObjects\"
	SingleInstanceEventName = "DiabloII Check For Other Instances"
)
