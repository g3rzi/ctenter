package discover

import (
	"fmt"
	"log"
	"sort"
)

// Runtime selects which container runtime provider to use.
type Runtime string

const (
	RuntimeAuto   Runtime = "auto"   // try all available providers
	RuntimeDocker Runtime = "docker" // force Docker only
	RuntimeCRI    Runtime = "cri"    // force CRI (crictl) only
)

// ParseRuntime validates and returns a Runtime value.
func ParseRuntime(s string) (Runtime, error) {
	switch Runtime(s) {
	case RuntimeAuto, RuntimeDocker, RuntimeCRI:
		return Runtime(s), nil
	default:
		return "", fmt.Errorf("unknown runtime %q: must be one of auto, docker, cri", s)
	}
}

type Discoverer struct {
	verbose   bool
	providers []Provider
}

func New(verbose bool) *Discoverer {
	return NewWithRuntime(verbose, RuntimeAuto)
}

func NewWithRuntime(verbose bool, runtime Runtime) *Discoverer {
	d := &Discoverer{verbose: verbose}

	switch runtime {
	case RuntimeDocker:
		d.providers = []Provider{NewDockerProvider(verbose)}
	case RuntimeCRI:
		d.providers = []Provider{NewCRIProvider(verbose)}
	default:
		// auto: try CRI first, then Docker
		d.providers = []Provider{
			NewCRIProvider(verbose),
			NewDockerProvider(verbose),
		}
	}

	return d
}

func (d *Discoverer) ListContainers() ([]*Container, error) {
	var allContainers []*Container
	seen := make(map[int]bool)
	
	for _, provider := range d.providers {
		if !provider.IsAvailable() {
			if d.verbose {
				log.Printf("Provider %s not available, skipping", provider.Name())
			}
			continue
		}
		
		if d.verbose {
			log.Printf("Discovering containers using %s provider", provider.Name())
		}
		
		containers, err := provider.Discover()
		if err != nil {
			if d.verbose {
				log.Printf("Provider %s failed: %v", provider.Name(), err)
			}
			continue
		}
		
		// Deduplicate by PID
		for _, container := range containers {
			if !seen[container.PID] {
				allContainers = append(allContainers, container)
				seen[container.PID] = true
			}
		}
		
		if d.verbose {
			log.Printf("Provider %s found %d containers", provider.Name(), len(containers))
		}
	}
	
	// Sort by PID
	sort.Slice(allContainers, func(i, j int) bool {
		return allContainers[i].PID < allContainers[j].PID
	})
	
	return allContainers, nil
}