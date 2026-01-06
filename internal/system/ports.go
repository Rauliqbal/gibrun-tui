package system

import (
	"os/exec"
	"regexp"
	"strings"
)

func GetPortUsage(svcName string) string {
	cmd := exec.Command("pgrep", svcName)
	pid, err := cmd.Output()
	if err != nil || len(pid) == 0 {
		return "N/A"
	}

	ssCmd := exec.Command("sudo", "ss", "-tulpn")
	output, _ := ssCmd.Output()

	cleanPid := strings.TrimSpace(string(pid))
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, "pid="+cleanPid) {
			re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+:(\d+)|\[::\]:(\d+)`)
			match := re.FindStringSubmatch(line)
			if len(match) > 0 {
				return match[0]
			}
		}
	}
	return "Scanning..."
}
