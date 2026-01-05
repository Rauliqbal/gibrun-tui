package system

import "github.com/Rauliqbal/gibrun/internal/config"

func ResolveService(
	cfg config.Services,
	key string,
) string {
	distro := DetectDistro()

	if s, ok := cfg[key]; ok {
		if name, ok := s.Services[distro]; ok {
			return name
		}
	}
	return ""
}
