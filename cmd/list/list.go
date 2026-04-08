package list

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/g3rzi/ctenter/pkg/discover"
)

var (
	verboseFlag bool
	wideFlag    bool
	noTrunc     bool
	runtimeFlag string
)

func NewListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all containers on the machine",
		Long:  `List all containers on the machine with PID, pod name, container name, and namespace.`,
		Run:   runList,
	}

	listCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")
	listCmd.Flags().BoolVarP(&wideFlag, "wide", "w", false, "show additional columns (pod id, container id, image id)")
	listCmd.Flags().BoolVar(&noTrunc, "no-trunc", false, "do not truncate long fields")
	listCmd.Flags().StringVarP(&runtimeFlag, "runtime", "r", "auto", "container runtime to query: auto, docker, cri")

	return listCmd
}

func runList(cmd *cobra.Command, args []string) {
	if os.Geteuid() != 0 {
		log.Fatalf("ctenter requires root privileges")
	}

	rt, err := discover.ParseRuntime(runtimeFlag)
	if err != nil {
		log.Fatalf("Invalid --runtime flag: %v", err)
	}

	discoverer := discover.NewWithRuntime(verboseFlag, rt)
	containers, err := discoverer.ListContainers()
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	if wideFlag {
		// Wide view: PID, podname, container name, namespace, pod id, container id, image id
		fmt.Printf("%-8s %-20s %-20s %-15s %-15s %-15s %-20s\n",
			"PID", "POD", "CONTAINER", "NAMESPACE", "POD_ID", "CONTAINER_ID", "IMAGE_ID")
		fmt.Printf("%-8s %-20s %-20s %-15s %-15s %-15s %-20s\n",
			"---", "---", "---------", "---------", "------", "------------", "--------")

		for _, c := range containers {
			podName := getDisplayValue(c.PodName, "N/A")
			containerName := getDisplayValue(c.ContainerName, "N/A")
			namespace := getDisplayValue(c.Namespace, "N/A")
			podID := short(c.PodID, 12)
			containerID := short(c.ContainerID, 12)
			imageID := short(c.ImageID, 18)

			fmt.Printf("%-8d %-20s %-20s %-15s %-15s %-15s %-20s\n",
				c.PID,
				maybeTrunc(podName, 20),
				maybeTrunc(containerName, 20),
				maybeTrunc(namespace, 15),
				maybeTrunc(podID, 15),
				maybeTrunc(containerID, 15),
				maybeTrunc(imageID, 20),
			)
		}
	} else {
		// Default view: PID, podname, container name, namespace
		fmt.Printf("%-8s %-25s %-25s %-15s\n",
			"PID", "POD", "CONTAINER", "NAMESPACE")
		fmt.Printf("%-8s %-25s %-25s %-15s\n",
			"---", "---", "---------", "---------")

		for _, c := range containers {
			podName := getDisplayValue(c.PodName, "N/A")
			containerName := getDisplayValue(c.ContainerName, "N/A")
			namespace := getDisplayValue(c.Namespace, "N/A")

			fmt.Printf("%-8d %-25s %-25s %-15s\n",
				c.PID,
				maybeTrunc(podName, 25),
				maybeTrunc(containerName, 25),
				maybeTrunc(namespace, 15),
			)
		}
	}

	fmt.Printf("\nFound %d containers\n", len(containers))
}

// getDisplayValue returns the value if not empty, otherwise returns the fallback
func getDisplayValue(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func short(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func maybeTrunc(s string, width int) string {
	if noTrunc || width <= 0 || len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}