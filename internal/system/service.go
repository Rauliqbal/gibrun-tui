package system

import (
	"os/exec"
	"strings"
)

func run(args ...string) string {
	out, _ := exec.Command(args[0], args[1:]...).Output()
	return strings.TrimSpace(string(out))
}

func IsRunning(service string) bool {
	return run("systemctl", "is-active", service) == "active"
}

func Start(s string) {
	exec.Command("pkexec", "systemctl", "start", s).Run()
}

func Stop(s string) {
	exec.Command("pkexec", "systemctl", "stop", s).Run()
}

func Restart(s string) {
	exec.Command("pkexec", "systemctl", "restart", s).Run()
}
