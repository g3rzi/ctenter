package nsenter

import (
	// "bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	// "strings"
	"syscall"
	"golang.org/x/sys/unix"
)

type NSEnter struct {
	verbose bool
}

func New(verbose bool) *NSEnter {
	return &NSEnter{verbose: verbose}
}

func (n *NSEnter) ExecCommand(pid int, agentPath string, args []string) (string, error) {
	if n.verbose {
		log.Printf("Executing command in container PID %d: %s %v", pid, agentPath, args)
	}
	
	// Try native setns first, fallback to nsenter
	if err := n.tryNativeSetNS(pid, agentPath, args); err == nil {
		return "", nil // Native method handles output directly
	}
	
	// Fallback to nsenter command
	return n.execWithNSEnter(pid, agentPath, args)
}

func (n *NSEnter) InteractiveShell(pid int, agentPath string) error {
	if n.verbose {
		log.Printf("Starting interactive shell in container PID %d", pid)
	}
	
	// Try native setns first, fallback to nsenter
	if err := n.tryNativeInteractive(pid, agentPath); err == nil {
		return nil
	}
	
	// Fallback to nsenter command
	return n.interactiveWithNSEnter(pid, agentPath)
}

func (n *NSEnter) tryNativeSetNS(pid int, agentPath string, args []string) error {
	// TODO: Implement native setns - complex due to Go runtime limitations
	// For now, return error to trigger fallback
	return fmt.Errorf("native setns not implemented yet")
}

func (n *NSEnter) tryNativeInteractive(pid int, agentPath string) error {
	// TODO: Implement native interactive setns
	return fmt.Errorf("native interactive setns not implemented yet")
}

func (n *NSEnter) execWithNSEnter(pid int, agentPath string, args []string) (string, error) {
	// Build nsenter command
	cmdArgs := []string{
		"-t", fmt.Sprintf("%d", pid),
		"-a", // Enter all namespaces
		agentPath,
	}
	cmdArgs = append(cmdArgs, args...)
	
	cmd := exec.Command("nsenter", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("nsenter failed: %v", err)
	}
	
	return string(output), nil
}

func (n *NSEnter) interactiveWithNSEnter(pid int, agentPath string) error {
	// Build nsenter command for interactive shell
	cmdArgs := []string{
		"-t", fmt.Sprintf("%d", pid),
		"-a", // Enter all namespaces
		agentPath, // Run agent in interactive mode (no args)
	}
	
	cmd := exec.Command("nsenter", cmdArgs...)
	
	// Connect stdin/stdout/stderr for interactive experience
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Handle signals properly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()
	
	if n.verbose {
		log.Printf("Starting nsenter: %v", cmdArgs)
	}
	
	return cmd.Run()
}

// Future native implementation helpers (for reference)
func (n *NSEnter) openNamespaces(pid int) (map[string]*os.File, error) {
	namespaces := []string{"mnt", "uts", "ipc", "net", "pid", "user"}
	nsFiles := make(map[string]*os.File)
	
	for _, ns := range namespaces {
		nsPath := fmt.Sprintf("/proc/%d/ns/%s", pid, ns)
		file, err := os.Open(nsPath)
		if err != nil {
			// Close any already opened files
			for _, f := range nsFiles {
				f.Close()
			}
			return nil, fmt.Errorf("failed to open %s namespace: %v", ns, err)
		}
		nsFiles[ns] = file
	}
	
	return nsFiles, nil
}

func (n *NSEnter) enterNamespaces(nsFiles map[string]*os.File) error {
	// Order matters for some namespaces
	order := []string{"user", "mnt", "uts", "ipc", "net", "pid"}
	
	for _, ns := range order {
		if file, exists := nsFiles[ns]; exists {
			if err := unix.Setns(int(file.Fd()), 0); err != nil {
				return fmt.Errorf("failed to enter %s namespace: %v", ns, err)
			}
		}
	}
	
	return nil
}
