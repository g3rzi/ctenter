package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/g3rzi/ctenter/cmd/list"
	"path/filepath"
	"github.com/g3rzi/ctenter/pkg/inject"
	"github.com/g3rzi/ctenter/pkg/nsenter"
)

var (
	version   = "v0.1.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

var (
	pidFlag       int
	execFlag      string
	agentPathFlag string
	verboseFlag   bool
)

var rootCmd = &cobra.Command{
	Use:   "ctenter",
	Short: "Container shell tool for listing and accessing containers",
	Long: `ctenter is a host-side tool that can list containers across different runtimes
(Docker, containerd, CRI-O) and inject agents to get shell access.`,
	Version: version,
	Run:     runRoot,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")

	// Add shell command flags to root for backward compatibility
	rootCmd.Flags().IntVarP(&pidFlag, "pid", "p", 0, "target container PID")
	rootCmd.Flags().StringVarP(&execFlag, "exec", "e", "", "execute command instead of interactive shell")
	rootCmd.Flags().StringVar(&agentPathFlag, "agent-path", "", "path to custom agent binary (defaults to embedded ctenterd)")

	// Add subcommands
	rootCmd.AddCommand(list.NewListCmd())
	rootCmd.AddCommand(newShellCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	if pidFlag != 0 {
		runShell(cmd, args)
	} else {
		cmd.Help()
	}
}

func newShellCmd() *cobra.Command {
	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Get shell access to a container",
		Run:   runShell,
	}
	shellCmd.Flags().IntVarP(&pidFlag, "pid", "p", 0, "target container PID (required)")
	shellCmd.Flags().StringVarP(&execFlag, "exec", "e", "", "execute command instead of interactive shell")
	shellCmd.Flags().StringVar(&agentPathFlag, "agent-path", "", "path to custom agent binary (defaults to embedded ctenterd)")
	shellCmd.MarkFlagRequired("pid")

	return shellCmd
}

func runShell(cmd *cobra.Command, args []string) {
	if os.Geteuid() != 0 {
		log.Fatalf("ctenter requires root privileges")
	}

	if pidFlag == 0 {
		log.Fatalf("--pid flag is required")
	}

	// Validate PID exists
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", pidFlag)); os.IsNotExist(err) {
		log.Fatalf("Process %d does not exist", pidFlag)
	}

	var agentBytes []byte
	var agentName string
	var err error

	if agentPathFlag != "" {
		agentBytes, err = os.ReadFile(agentPathFlag)
		if err != nil {
			log.Fatalf("Failed to read agent binary %s: %v", agentPathFlag, err)
		}
		agentName = "custom-agent"
	} else {
		agentBytes, err = getEmbeddedAgent()
		if err != nil {
			log.Fatalf("Failed to get embedded agent: %v", err)
		}
		agentName = "ctenterd"
	}

	if verboseFlag {
		log.Printf("Using agent: %s (%d bytes)", agentName, len(agentBytes))
	}

	// Inject agent
	injector := inject.New(verboseFlag)
	agentPath := fmt.Sprintf("/tmp/%s", agentName)
	
	if err := injector.InjectAgent(pidFlag, agentPath, agentBytes); err != nil {
		log.Fatalf("Failed to inject agent: %v", err)
	}

	if verboseFlag {
		log.Printf("Agent injected to /proc/%d/root%s", pidFlag, agentPath)
	}

	// Enter container namespaces and execute
	nsEnter := nsenter.New(verboseFlag)
	
	if execFlag != "" {
		// One-shot execution
		result, err := nsEnter.ExecCommand(pidFlag, agentPath, []string{execFlag})
		if err != nil {
			log.Fatalf("Failed to execute command: %v", err)
		}
		fmt.Print(result)
	} else {
		// Interactive shell
		if err := nsEnter.InteractiveShell(pidFlag, agentPath); err != nil {
			log.Fatalf("Failed to start interactive shell: %v", err)
		}
	}
}

func getEmbeddedAgent() ([]byte, error) {
	// In a real implementation, you'd embed the ctenterd binary
	// For now, try to find it in the same directory or PATH
	
	// Try current directory first
	if data, err := os.ReadFile("bin/ctenterd"); err == nil {
		return data, nil
	}
	
	// Try looking for ctenterd in PATH
	if data, err := os.ReadFile("/usr/local/bin/ctenterd"); err == nil {
		return data, nil
	}
	
	// Try relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if data, err := os.ReadFile(filepath.Join(dir, "ctenterd")); err == nil {
			return data, nil
		}
	}
	
	return nil, fmt.Errorf("ctenterd binary not found - build it first with 'make ctenterd'")
}
