package shell

import (
	"github.com/g3rzi/ctenter/pkg/nsenter"
)

type ExecShell struct {
	verbose bool
	nsenter *nsenter.NSEnter
}

func NewExec(verbose bool) *ExecShell {
	return &ExecShell{
		verbose: verbose,
		nsenter: nsenter.New(verbose),
	}
}

func (e *ExecShell) Execute(pid int, agentPath string, command string) (string, error) {
	args := []string{command}
	return e.nsenter.ExecCommand(pid, agentPath, args)
}
