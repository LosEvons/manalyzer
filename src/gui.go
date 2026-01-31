package manalyzer

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App          *tview.Application
	Pages        *tview.Pages
	Root         *tview.Flex
	Left         *tview.Form
	Middle       *tview.TextView
	Bottom       *tview.TextView
	BasePathField *tview.InputField
	SteamIDFields [5]*tview.InputField
	AnalyzeButton *tview.Button
}

func New() *UI {
	app := tview.NewApplication()

	// Create UI components  
	middle := tview.NewTextView()
	middle.SetBorder(true).SetTitle("Status")
	middle.SetText("Ready. Enter base path and steamID64s, then click Analyze.")
	
	bottom := tview.NewTextView()
	bottom.SetBorder(true).SetTitle("Statistics")
	bottom.SetScrollable(true)
	
	ui := &UI{
		App:    app,
		Middle: middle,
		Bottom: bottom,
	}

	// Create form for left panel
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Configuration")

	// Base path input field
	ui.BasePathField = tview.NewInputField().
		SetLabel("Base Path: ").
		SetFieldWidth(30).
		SetPlaceholder("/path/to/demos")
	form.AddFormItem(ui.BasePathField)

	// Add 5 steamID64 input fields
	for i := 0; i < 5; i++ {
		field := tview.NewInputField().
			SetLabel(fmt.Sprintf("SteamID64 #%d: ", i+1)).
			SetFieldWidth(20).
			SetPlaceholder("76561198...")
		ui.SteamIDFields[i] = field
		form.AddFormItem(field)
	}

	// Add analyze button
	form.AddButton("Analyze", func() {
		ui.onAnalyze()
	})

	form.AddButton("Quit", func() {
		app.Stop()
	})

	ui.Left = form

	// Create layout
	col := tview.NewFlex().
		AddItem(ui.Left, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(ui.Middle, 0, 3, false).
			AddItem(ui.Bottom, 10, 1, false), 0, 2, false)

	pages := tview.NewPages().AddPage("main", col, true, true)

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC, tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
	})

	ui.Pages = pages
	ui.Root = col

	return ui
}

func (u *UI) Start() error {
	return u.App.Run()
}

func (u *UI) Stop() {
	u.App.Stop()
}

func (u *UI) QueueUpdate(fn func()) {
	u.App.QueueUpdateDraw(fn)
}

func (u *UI) SetMiddleText(text string) {
	u.QueueUpdate(func() {
		u.Middle.Clear()
		fmt.Fprint(u.Middle, text)
	})
}

func (u *UI) SetBottomText(text string) {
	u.QueueUpdate(func() {
		u.Bottom.Clear()
		fmt.Fprint(u.Bottom, text)
		u.Bottom.ScrollToBeginning()
	})
}

// onAnalyze is called when the Analyze button is clicked
func (u *UI) onAnalyze() {
	// Get base path
	basePath := strings.TrimSpace(u.BasePathField.GetText())
	if basePath == "" {
		u.SetMiddleText("[red]Error:[white] Base path is required")
		return
	}

	// Get steamID64s
	var steamIDs []uint64
	for i, field := range u.SteamIDFields {
		text := strings.TrimSpace(field.GetText())
		if text == "" {
			continue // Skip empty fields
		}

		steamID, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			u.SetMiddleText(fmt.Sprintf("[red]Error:[white] Invalid SteamID64 #%d: %s", i+1, text))
			return
		}
		steamIDs = append(steamIDs, steamID)
	}

	if len(steamIDs) == 0 {
		u.SetMiddleText("[red]Error:[white] At least one SteamID64 is required")
		return
	}

	// Update status
	u.SetMiddleText(fmt.Sprintf("[yellow]Analyzing demos from %s...[white]", basePath))
	u.SetBottomText("Processing...")

	// Run analysis in goroutine to keep UI responsive
	go func() {
		// Gather all demos
		matches, err := GatherAllDemos(basePath)
		if err != nil {
			if err == ErrNoDemos {
				u.SetMiddleText(fmt.Sprintf("[red]Error:[white] No .dem files found in %s", basePath))
				u.SetBottomText("")
				return
			}
			log.Printf("Error gathering demos: %v", err)
			u.SetMiddleText(fmt.Sprintf("[red]Error:[white] Failed to gather demos: %v", err))
			u.SetBottomText("")
			return
		}

		// Filter and group player stats
		stats, err := FilterAndGroupPlayerStats(matches, steamIDs)
		if err != nil {
			if err == ErrNoValidSteamIDs {
				u.SetMiddleText("[red]Error:[white] None of the provided SteamID64s were found in any demos")
				u.SetBottomText("")
				return
			}
			log.Printf("Error filtering stats: %v", err)
			u.SetMiddleText(fmt.Sprintf("[red]Error:[white] Failed to filter stats: %v", err))
			u.SetBottomText("")
			return
		}

		// Display results
		statusMsg := fmt.Sprintf("[green]Success![white] Analyzed %d match(es)", len(stats.ByMatch))
		if len(stats.NotFound) > 0 {
			statusMsg += fmt.Sprintf("\n[yellow]Warning:[white] %d SteamID64(s) not found", len(stats.NotFound))
		}
		u.SetMiddleText(statusMsg)

		// Format and display stats
		formattedStats := FormatStatsForDisplay(stats)
		u.SetBottomText(formattedStats)
	}()
}
