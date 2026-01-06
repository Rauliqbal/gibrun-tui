package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Rauliqbal/gibrun/internal/ui"
)

func main() {
	versionFlag := flag.Bool("version", false, "Tampilkan versi aplikasi")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("GibRun version v%s\n", ui.Version)
		os.Exit(0)
	}

	ui.Run()
}
