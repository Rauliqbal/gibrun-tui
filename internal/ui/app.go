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

// VERSION
var Version = "0.1.0"

// Helper untuk membersihkan label list agar mendapatkan key service murni
func cleanServiceName(text string) string {
	re := regexp.MustCompile(`\[[^\]]*\]`)
	clean := re.ReplaceAllString(text, "")
	clean = strings.ReplaceAll(clean, "●", "")
	clean = strings.ReplaceAll(clean, "○", "")
	return strings.TrimSpace(clean)
}

// Merender teks untuk item di list service
func renderService(key string, cfg config.Services) string {
	svc := system.ResolveService(cfg, key)
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

	var keys []string
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		label := renderService(key, cfg)
		list.AddItem(label, "", 0, nil)
	}

	if currentIndex >= 0 && currentIndex < list.GetItemCount() {
		list.SetCurrentItem(currentIndex)
	}
}

func Run() {
	app := tview.NewApplication()
	cfg, err := config.Load("internal/config/services.yml")
	if err != nil {
		panic(err)
	}

	// State Management untuk Log
	var cancelLog context.CancelFunc

	// UI Components
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	distro := system.DetectDistro()
	header.SetText(fmt.Sprintf("[yellow]⚡ GibRun[white] — Distro: [blue]%s", distro))

	serviceList := tview.NewList().ShowSecondaryText(false)
	serviceList.SetSelectedBackgroundColor(tcell.ColorDeepSkyBlue)
	serviceList.SetBorder(true).SetTitle(" Services ")

	logView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	logView.SetBorder(true).SetTitle(" Realtime Logs & Info ")

	// Footer
	footerMenu := tview.NewTextView().SetDynamicColors(true)
	footerMenu.SetText("[::b]S[white]: Start  [::b]R[white]: Restart  [::b]X[white]: Stop  [::b]I[white]: Install  [::b]Q[white]: Quit")

	footerVersion := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	footerVersion.SetText(fmt.Sprintf("[gray]v%s", Version))

	footerWrapper := tview.NewFlex().AddItem(footerMenu, 0, 1, false).AddItem(footerVersion, 15, 0, false)

	// Main Logic: Update Info & Streaming Log
	updateInfo := func() {
		// 1. Cancel log goroutine sebelumnya
		if cancelLog != nil {
			cancelLog()
		}

		idx := serviceList.GetCurrentItem()
		if idx < 0 {
			return
		}

		fullText, _ := serviceList.GetItemText(idx)
		key := cleanServiceName(fullText)
		svcName := system.ResolveService(cfg, key)

		logView.Clear()

		if !system.IsInstalled(svcName) {
			fmt.Fprintf(logView, "\n [red]Service '%s' belum terinstall di %s.\n [white]Tekan 'I' untuk instalasi.", key, distro)
			return
		}

		// 2. Tampilkan Meta Info
		port := system.GetPortUsage(svcName)
		uptime := system.GetUptime(svcName)
		fmt.Fprintf(logView, "[yellow]Service:[white] %s | [yellow]Port:[white] %s | [yellow]Uptime:[white] %s\n%s\n",
			key, port, uptime, strings.Repeat("─", 50))

		// 3. Start New Log Stream dengan Context
		var ctx context.Context
		ctx, cancelLog = context.WithCancel(context.Background())

		go func() {
			logChan := system.StreamLogs(ctx, svcName)
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-logChan:
					if !ok {
						return
					}
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logView, "%s\n", line)
					})
				}
			}
		}()
	}

	// Event Handlers
	serviceList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		updateInfo()
	})

	serviceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := serviceList.GetCurrentItem()
		if idx < 0 {
			return event
		}

		fullText, _ := serviceList.GetItemText(idx)
		key := cleanServiceName(fullText)
		svc := system.ResolveService(cfg, key)

		switch event.Rune() {
		case 'q', 'Q':
			app.Stop()
		case 'i', 'I':
			logView.SetText(fmt.Sprintf("[yellow]Installing %s...", key))
			go func() {
				output, err := system.InstallService(svc)
				app.QueueUpdateDraw(func() {
					if err != nil {
						fmt.Fprintf(logView, "\n[red]Error: %v\n%s", err, string(output))
					} else {
						fmt.Fprintf(logView, "\n[green]Sukses!\n%s", string(output))
						refreshServices(serviceList, cfg)
					}
				})
			}()
		case 's', 'S', 'r', 'R', 'x', 'X':
			go func() {
				switch event.Rune() {
				case 's', 'S':
					system.Start(svc)
				case 'r', 'R':
					system.Restart(svc)
				case 'x', 'X':
					system.Stop(svc)
				}
				app.QueueUpdateDraw(func() {
					refreshServices(serviceList, cfg)
					updateInfo()
				})
			}()
		}
		return event
	})

	// Initial Load
	refreshServices(serviceList, cfg)
	updateInfo()

	// Layouting
	mainFlex := tview.NewFlex().AddItem(serviceList, 0, 1, true).AddItem(logView, 0, 2, false)
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(mainFlex, 0, 1, true).
		AddItem(footerWrapper, 1, 0, false)

	// Final Cleanup saat aplikasi exit
	defer func() {
		if cancelLog != nil {
			cancelLog()
		}
	}()

	if err := app.SetRoot(root, true).SetFocus(serviceList).Run(); err != nil {
		panic(err)
	}
}
