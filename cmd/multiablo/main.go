package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/chenwei791129/multiablo/internal/handle"
	"github.com/chenwei791129/multiablo/internal/process"
	"github.com/chenwei791129/multiablo/pkg/d2r"
)

var (
	verbose bool
	logger  *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "multiablo",
	Short: "D2R Multi-Instance Helper",
	Long: `Multiablo enables you to run multiple instances of Diablo II: Resurrected
simultaneously by continuously monitoring and removing the "DiabloII Check For Other Instances" and "Agent.exe".`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger based on verbose flag
		var err error
		if verbose {
			// Development mode with console encoder for better readability
			config := zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			logger, err = config.Build()
		} else {
			// Production mode with custom config for cleaner output
			config := zap.NewProductionConfig()

			config.DisableCaller = true
			config.DisableStacktrace = true
			config.Encoding = "console"
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			logger, err = config.Build()
		}
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			_ = logger.Sync()
		}()

		logger.Info("Multiablo - D2R Multi-Instance Helper")
		logger.Info("======================================")
		logger.Info("")

		// Start background goroutines
		stopChan := make(chan struct{})

		logger.Info("Starting background monitors...")
		go handleCloserLoop(stopChan)
		go agentKillerLoop(stopChan)

		// Give goroutines time to print their startup messages
		time.Sleep(10 * time.Millisecond)

		logger.Info("")
		logger.Info("Press Enter to exit...")

		// Wait for user to press Enter
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		close(stopChan)

		// Give goroutines time to finish
		time.Sleep(100 * time.Millisecond)
		logger.Info("Stopped.")

		return nil
	},
}

func init() {
	// Disable Cobra's mousetrap feature on Windows
	// By default, Cobra shows a warning when launched from File Explorer instead of cmd.exe
	// Setting this to empty string allows the program to run normally from File Explorer
	// See: https://github.com/spf13/cobra/issues/844
	cobra.MousetrapHelpText = ""
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// handleCloserLoop continuously monitors D2R processes and closes their "single-instance" handles
func handleCloserLoop(stopChan chan struct{}) {
	ticker := time.NewTicker(1 * time.Second) // Check every 1 second
	defer ticker.Stop()

	logger.Info("Monitoring " + d2r.ProcessName + " processes for handle restrictions...")

	for {
		select {
		case <-stopChan:
			logger.Info("Stopping handle monitor...")
			return
		case <-ticker.C:
			// Find all D2R processes
			processes, err := process.FindProcessesByName(d2r.ProcessName)
			if err != nil {
				logger.Debug("Error finding D2R processes", zap.Error(err))
				continue
			}

			// Close handles for each D2R process
			for _, proc := range processes {
				closedCount, err := handle.CloseHandlesByName(proc.PID, d2r.SingleInstanceEventName)
				if err != nil {
					// Ignore "no handles found" errors as they're expected after first close
					continue
				}
				if closedCount > 0 {
					logger.Info("Removed multi-instance restriction",
						zap.Uint32("pid", proc.PID),
						zap.Int("handles_closed", closedCount))
				}
			}
		}
	}
}

// agentKillerLoop continuously monitors and kills Agent.exe processes
func agentKillerLoop(stopChan chan struct{}) {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 second
	defer ticker.Stop()

	logger.Info("Monitoring " + d2r.AgentProcessName + " processes for termination...")

	for {
		select {
		case <-stopChan:
			logger.Info("Stopping Agent.exe monitor...")
			return
		case <-ticker.C:
			// Check if Agent.exe is running
			running, err := process.IsProcessRunning(d2r.AgentProcessName)
			if err != nil {
				logger.Debug("Error checking Agent.exe", zap.Error(err))
				continue
			}

			if running {
				// Kill all Agent.exe processes
				killedCount, err := process.KillProcessesByName(d2r.AgentProcessName)
				if err != nil {
					logger.Debug("Failed to kill Agent.exe", zap.Error(err))
					continue
				}
				if killedCount > 0 {
					logger.Info("Terminated "+d2r.AgentProcessName+" processes", zap.Int("count", killedCount))
				}
			}
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
