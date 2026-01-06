package system

import (
	"os/exec"
	"strings"
	"time"
)

func run(args ...string) string {
	out, _ := exec.Command(args[0], args[1:]...).Output()
	return strings.TrimSpace(string(out))
}

func IsRunning(service string) bool {
	return run("systemctl", "is-active", service) == "active"
}

func Start(s string) {
	exec.Command("systemctl", "start", s).Run()
}

func Stop(s string) {
	exec.Command("systemctl", "stop", s).Run()
}

func Restart(s string) {
	exec.Command("systemctl", "restart", s).Run()
}

func ServiceExists(service string) bool {
	return run("systemctl", "list-unit-files", service+".service") != ""
}

func IsInstalled(svcName string) bool {
	if svcName == "" {
		return false
	}

	cmd := exec.Command("systemctl", "list-unit-files", svcName+".service")
	output, _ := cmd.Output()
	if strings.Contains(string(output), svcName+".service") {
		return true
	}

	_, err := exec.LookPath(svcName)
	return err == nil
}

func GetUptime(svcName string) string {
	cmd := exec.Command("systemctl", "show", svcName, "--property=ActiveEnterTimestamp")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}

	// Format output: ActiveEnterTimestamp=Tue 2024-05-21 10:00:00 WIB
	parts := strings.Split(string(output), "=")
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return "Stopped"
	}

	layout := "Mon 2006-01-02 15:04:05 MST"
	t, err := time.Parse(layout, strings.TrimSpace(parts[1]))
	if err != nil {
		return "Active"
	}

	duration := time.Since(t)
	return duration.Round(time.Second).String()
}
