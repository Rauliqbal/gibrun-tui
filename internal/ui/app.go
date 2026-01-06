package ui

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Rauliqbal/gibrun/internal/config"
	"github.com/Rauliqbal/gibrun/internal/system"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func cleanServiceName(text string) string {
	re := regexp.MustCompile(`\[[^\]]*\]`)
	clean := re.ReplaceAllString(text, "")
	clean = strings.ReplaceAll(clean, "●", "")
	clean = strings.ReplaceAll(clean, "○", "")
	return strings.TrimSpace(clean)
}

func renderService(key string, cfg config.Services) string {
	svc := system.ResolveService(cfg, key)
	// Check if installed
	if !system.IsInstalled(svc) {
		return "[gray]○ [darkgray]" + key + " [red](Not Installed)"
	}
	if system.IsRunning(svc) {
		return "[green]● [white]" + key
	}
	return "[red]○ [white]" + key
}

func refreshServices(list *tview.List, cfg config.Services) {
	currentIndex := list.GetCurrentItem()
	list.Clear()

	if cfg == nil || len(cfg) == 0 {
		return
	}

	var keys []string
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		// Pastikan renderService mengembalikan string yang valid
		label := renderService(key, cfg)
		list.AddItem(label, "", 0, nil)
	}

	if currentIndex >= 0 && currentIndex < list.GetItemCount() {
		list.SetCurrentItem(currentIndex)
	} else if list.GetItemCount() > 0 {
		list.SetCurrentItem(0)
	}
}

func Run() {
	app := tview.NewApplication()
	cfg, _ := config.Load("internal/config/services.yml")

	stopLogChan := make(chan struct{}, 1)

	if len(cfg) == 0 {
		fmt.Println("DEBUG: Config terbaca tapi kosong!")
	}

	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	distroName := system.DetectDistro()
	header.SetText(fmt.Sprintf("[yellow]⚡ GibRun[white] — Distro: [blue]%s", distroName))

	serviceList := tview.NewList().ShowSecondaryText(false)
	serviceList.SetBorder(true).SetTitle(" Services ")
	serviceList.SetSelectedBackgroundColor(tcell.ColorDeepSkyBlue)

	logView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	logView.SetBorder(true).SetTitle(" Realtime Logs & Info ")

	footer := tview.NewTextView().SetDynamicColors(true)
	footer.SetText("[::b]S[white]: Start  [::b]R[white]: Restart  [::b]X[white]: Stop  [::b]I[white]: Install  [::b]Q[white]: Quit")

	var cancelLog context.CancelFunc

	// Fungsi untuk memantau log secara realtime
	updateInfo := func() {
		if cancelLog != nil {
			cancelLog()
		}

		select {
		case stopLogChan <- struct{}{}:
		default:
		}

		if serviceList.GetItemCount() == 0 {
			logView.SetText("No services found in configuration.")
			return
		}

		idx := serviceList.GetCurrentItem()
		if idx < 0 || idx >= serviceList.GetItemCount() {
			return
		}

		fullText, _ := serviceList.GetItemText(idx)
		key := cleanServiceName(fullText)
		svcName := system.ResolveService(cfg, key)

		if key == "" {
			return
		}

		select {
		case stopLogChan <- struct{}{}:
		default:
		}

		var ctx context.Context
		ctx, cancelLog = context.WithCancel(context.Background())

		logView.Clear()

		if !system.IsInstalled(svcName) {
			logView.SetText("[red]Service not installed.")
			return
		}

		// Tampilkan Meta Info (Port & Uptime)
		port := system.GetPortUsage(svcName) // Dari ports.go
		uptime := system.GetUptime(svcName)  // Dari service.go / uptime logic
		statusHeader := fmt.Sprintf("[yellow]Service:[white] %s | [yellow]Port:[white] %s | [yellow]Uptime:[white] %s\n%s\n",
			key, port, uptime, strings.Repeat("─", 50))

		logView.SetText(statusHeader)

		// Stream Realtime Log (Dari logs.go)
		go func() {
			logChan := system.StreamLogs(ctx, svcName)
			for {
				select {
				case line, ok := <-logChan:
					if !ok {
						return
					}
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logView, "%s\n", line)
					})
				case <-stopLogChan: // Berhenti jika ada sinyal pindah service
					return
				}
			}
		}()
	}

	serviceList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		updateInfo()
	})

	serviceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := serviceList.GetCurrentItem()
		fullText, _ := serviceList.GetItemText(idx)
		key := cleanServiceName(fullText)
		svc := system.ResolveService(cfg, key)

		if serviceList.GetItemCount() == 0 {
			return event
		}

		if idx < 0 || idx >= serviceList.GetItemCount() {
			return event
		}

		switch event.Rune() {
		case 'q', 'Q':
			app.Stop()
		case 'i', 'I':
			logView.SetText(fmt.Sprintf("[yellow]Installing %s...\n", key))
			go func() {
				output, err := system.InstallService(svc)

				app.QueueUpdateDraw(func() {
					if err != nil {
						// Tampilkan pesan error dan output terminal agar user tahu kenapa gagal
						fmt.Fprintf(logView, "[red]Gagal install: %v\n[white]%s", err, string(output))
					} else {
						fmt.Fprintf(logView, "[green]Instalasi %s berhasil!\n[white]%s", key, string(output))
						refreshServices(serviceList, cfg)
						updateInfo()
					}
				})
			}()
			return nil
		case 's', 'S':
			go func() {
				system.Start(svc)
				app.QueueUpdateDraw(func() { refreshServices(serviceList, cfg); updateInfo() })
			}()
		case 'r', 'R':
			go func() {
				system.Restart(svc)
				app.QueueUpdateDraw(func() { refreshServices(serviceList, cfg); updateInfo() })
			}()
		case 'x', 'X':
			go func() {
				system.Stop(svc)
				app.QueueUpdateDraw(func() { refreshServices(serviceList, cfg); updateInfo() })
			}()
		}
		return event
	})

	refreshServices(serviceList, cfg)
	mainFlex := tview.NewFlex().AddItem(serviceList, 0, 1, true).AddItem(logView, 0, 2, false)
	root := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(header, 1, 0, false).AddItem(mainFlex, 0, 1, true).AddItem(footer, 1, 0, false)

	if err := app.SetRoot(root, true).SetFocus(serviceList).Run(); err != nil {
		panic(err)
	}
}
