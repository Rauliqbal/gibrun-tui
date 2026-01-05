package ui

import (
	"fmt"
	"regexp"
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
	if svc == "" {
		return "[gray]○ [white]" + key
	}
	if system.IsRunning(svc) {
		return "[green]● [white]" + key
	}
	return "[red]○ [white]" + key
}

func refreshServices(list *tview.List, cfg config.Services) {
	currentIndex := list.GetCurrentItem()
	list.Clear()
	for key := range cfg {
		list.AddItem(renderService(key, cfg), "", 0, nil)
	}
	if currentIndex < list.GetItemCount() {
		list.SetCurrentItem(currentIndex)
	}
}

func Run() {
	app := tview.NewApplication()

	cfg, err := config.Load("internal/config/services.yml")
	if err != nil {
		panic(err)
	}

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]⚡ GibRun[white] — TUI Service Manager")

	serviceList := tview.NewList().ShowSecondaryText(false)
	serviceList.SetBorder(true).SetTitle(" Services ")
	refreshServices(serviceList, cfg)

	logView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	logView.SetBorder(true).SetTitle(" Logs ")
	logView.SetText("Gunakan keyboard untuk aksi...")

	footer := tview.NewTextView().SetDynamicColors(true)
	footer.SetText("[::b]S[white]: Start  [::b]R[white]: Restart  [::b]X[white]: Stop  [::b]Q[white]: Quit")

	mainFlex := tview.NewFlex().
		AddItem(serviceList, 0, 1, true).
		AddItem(logView, 0, 2, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(header, 1, 0, false).
		AddItem(mainFlex, 0, 1, true).
		AddItem(footer, 1, 0, false)

	// HANDLER INPUT
	serviceList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		idx := serviceList.GetCurrentItem()
		if idx < 0 {
			return event
		}

		// AMBIL NAMA SERVICE DENGAN BENAR
		fullText, _ := serviceList.GetItemText(idx)
		serviceKey := cleanServiceName(fullText)
		resolvedSvc := system.ResolveService(cfg, serviceKey)

		// Eksekusi berdasarkan tombol
		switch event.Rune() {
		case 'q', 'Q':
			app.Stop()
		case 's', 'S':
			if resolvedSvc != "" {
				logView.SetText(fmt.Sprintf("[green]Starting %s...", serviceKey))
				go func() {
					system.Start(resolvedSvc)
					app.QueueUpdateDraw(func() {
						refreshServices(serviceList, cfg)
					})
				}()
			}
			return nil
		case 'r', 'R':
			if resolvedSvc != "" {
				logView.SetText(fmt.Sprintf("[yellow]Restarting %s...", serviceKey))
				go func() {
					system.Restart(resolvedSvc)
					app.QueueUpdateDraw(func() {
						refreshServices(serviceList, cfg)
					})
				}()
			}
			return nil
		case 'x', 'X':
			if resolvedSvc != "" {
				logView.SetText(fmt.Sprintf("[red]Stopping %s...", serviceKey))
				go func() {
					system.Stop(resolvedSvc)
					app.QueueUpdateDraw(func() {
						refreshServices(serviceList, cfg)
					})
				}()
			}
			return nil
		}

		return event
	})

	if err := app.SetRoot(root, true).SetFocus(serviceList).Run(); err != nil {
		panic(err)
	}
}
