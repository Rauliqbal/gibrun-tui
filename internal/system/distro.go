package system

import (
	"os"
	"os/exec"
	"strings"
)

func DetectDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	content := strings.ToLower(string(data))

	if strings.Contains(content, "endeavouros") || strings.Contains(content, "arch") {
		return "arch"
	} else if strings.Contains(content, "ubuntu") || strings.Contains(content, "debian") {
		return "debian"
	} else if strings.Contains(content, "fedora") {
		return "fedora"
	}
	return "Linux"
}

func InstallService(svcName string) ([]byte, error) {
	distro := DetectDistro()
	var cmd *exec.Cmd

	switch distro {
	case "Debian/Ubuntu":
		cmd = exec.Command("sudo", "apt-get", "install", "-y", svcName)
	case "Arch Linux":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", svcName)
	case "Fedora/RHEL":
		cmd = exec.Command("sudo", "dnf", "install", "-y", svcName)
	default:
		return []byte("Distro tidak didukung"), nil
	}
	return cmd.CombinedOutput()
}
