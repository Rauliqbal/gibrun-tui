package system

import "os/exec"

func Logs(service string) *exec.Cmd {
	return exec.Command("journalctl", "-u", service, "-n", "100", "--no-pager")
}
