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

		cmd := exec.CommandContext(ctx, "journalctl", "-u", svcName, "-f", "-n", "20")
		stdout, _ := cmd.StdoutPipe()
		if err := cmd.Start(); err != nil {
			return
		}

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

		<-ctx.Done()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	return logChan
}
