package inject

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type Injector struct {
	verbose bool
}

func New(verbose bool) *Injector {
	return &Injector{verbose: verbose}
}

func (i *Injector) InjectAgent(pid int, targetPath string, agentBytes []byte) error {
	// Construct path inside container's root filesystem
	containerPath := fmt.Sprintf("/proc/%d/root%s", pid, targetPath)
	
	if i.verbose {
		log.Printf("Injecting %d bytes to %s", len(agentBytes), containerPath)
	}
	
	// Ensure target directory exists
	targetDir := filepath.Dir(containerPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %v", targetDir, err)
	}
	
	// Write agent binary
	if err := ioutil.WriteFile(containerPath, agentBytes, 0755); err != nil {
		if i.isReadOnlyError(err) {
			return fmt.Errorf("container has read-only filesystem - cannot inject agent")
		}
		return fmt.Errorf("failed to write agent: %v", err)
	}
	
	// Verify the file was written and calculate checksum
	if err := i.verifyInjection(containerPath, agentBytes); err != nil {
		return err
	}
	
	return nil
}

func (i *Injector) verifyInjection(path string, expectedBytes []byte) error {
	// Check file exists and is executable
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("injected file not found: %v", err)
	}
	
	// Check permissions
	if stat.Mode()&0111 == 0 {
		return fmt.Errorf("injected file is not executable")
	}
	
	// Read back and verify size
	actualBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read back injected file: %v", err)
	}
	
	if len(actualBytes) != len(expectedBytes) {
		return fmt.Errorf("size mismatch: expected %d bytes, got %d", len(expectedBytes), len(actualBytes))
	}
	
	// Calculate and report checksum
	expectedHash := CalculateSHA256(expectedBytes)
	actualHash := CalculateSHA256(actualBytes)
	
	if expectedHash != actualHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}
	
	if i.verbose {
		log.Printf("Injection verified: %d bytes, SHA256: %s", len(actualBytes), actualHash)
	}
	
	return nil
}

func (i *Injector) isReadOnlyError(err error) bool {
	// Check if error indicates read-only filesystem
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			return errno == syscall.EROFS
		}
	}
	return false
}
