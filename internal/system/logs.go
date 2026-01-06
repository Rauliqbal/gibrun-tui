package system

import (
	"bufio"
	"context"
	"os/exec"
)

func StreamLogs(ctx context.Context, svcName string) <-chan string {
	logChan := make(chan string)

	go func() {
		defer close(logChan)

		// Menggunakan CommandContext agar proses journalctl mati saat context di-cancel
		cmd := exec.CommandContext(ctx, "journalctl", "-u", svcName, "-f", "-n", "20")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return
		}

		if err := cmd.Start(); err != nil {
			return
		}

		// Scan baris demi baris
		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				case logChan <- scanner.Text():
				}
			}
		}()

		// Tunggu context selesai (akibat cancelLog() di UI)
		<-ctx.Done()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	return logChan
}
