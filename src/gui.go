package manalyzer

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	eventLogHeight = 5 // Height in rows for the event log panel
)

// ============================================================================
// DATA STRUCTURES
// ============================================================================

type PlayerInput struct {
	Name      string // For display purposes only
	SteamID64 string // Only SteamID64 format supported (17 digits)
}

type AnalysisConfig struct {
	Players  [5]PlayerInput // Support 1-5 players (not all slots required)
	BasePath string
}

// ============================================================================
// UI COMPONENTS
// ============================================================================

type UI struct {
	App        *tview.Application
	Pages      *tview.Pages
	Root       *tview.Flex
	form       *tview.Form
	eventLog   *EventLog
	statsTable *StatisticsTable
}

type EventLog struct {
	textView *tview.TextView
	maxLines int
	lines    []string
}

type StatisticsTable struct {
	table      *tview.Table
	data       *WrangleResult
	filterMap  string // "" = all maps, or specific map name
	filterSide string // "" = all sides, "T", or "CT"
}

// ============================================================================
// EVENT LOG IMPLEMENTATION
// ============================================================================

func newEventLog(maxLines int) *EventLog {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	
	tv.SetBorder(true)
	tv.SetTitle("Event Log")

	tv.SetChangedFunc(func() {
		// Auto-scroll to bottom
		tv.ScrollToEnd()
	})

	return &EventLog{
		textView: tv,
		maxLines: maxLines,
		lines:    make([]string, 0, maxLines),
	}
}

func (el *EventLog) Log(message string) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[yellow]%s[-] %s", timestamp, message)

	el.lines = append(el.lines, line)

	// Keep only last maxLines
	if len(el.lines) > el.maxLines {
		el.lines = el.lines[len(el.lines)-el.maxLines:]
	}

	// Update display by building the full text content
	var builder strings.Builder
	for i, l := range el.lines {
		builder.WriteString(l)
		if i < len(el.lines)-1 {
			builder.WriteString("\n")
		}
	}
	el.textView.SetText(builder.String())
}

func (el *EventLog) LogError(message string) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[yellow]%s[-] [red]ERROR:[-] %s", timestamp, message)

	el.lines = append(el.lines, line)
	if len(el.lines) > el.maxLines {
		el.lines = el.lines[len(el.lines)-el.maxLines:]
	}

	// Update display by building the full text content
	var builder strings.Builder
	for i, l := range el.lines {
		builder.WriteString(l)
		if i < len(el.lines)-1 {
			builder.WriteString("\n")
		}
	}
	el.textView.SetText(builder.String())
}

// ============================================================================
// STATISTICS TABLE IMPLEMENTATION
// ============================================================================

func newStatisticsTable() *StatisticsTable {
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0). // Fix header row
		SetSelectable(true, false)
	
	table.SetBorder(true)
	table.SetTitle("Player Statistics")

	return &StatisticsTable{
		table:      table,
		filterMap:  "",
		filterSide: "",
	}
}

func (st *StatisticsTable) UpdateData(result *WrangleResult) {
	st.data = result
	st.renderTable()
}

func (st *StatisticsTable) renderTable() {
	st.table.Clear()

	// Header row with column names
	headers := []string{"Player", "Map", "Side", "KAST%", "ADR", "K/D",
		"Kills", "Deaths", "FK", "FD", "TK", "TD"}

	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)
		st.table.SetCell(0, col, cell)
	}

	// Data rows
	row := 1
	if st.data != nil {
		// Sort by player name initially
		sortedPlayers := make([]*PlayerStats, 0, len(st.data.PlayerStats))
		for _, ps := range st.data.PlayerStats {
			if ps != nil {
				sortedPlayers = append(sortedPlayers, ps)
			}
		}
		sort.Slice(sortedPlayers, func(i, j int) bool {
			// Defensive check for nil
			if sortedPlayers[i] == nil || sortedPlayers[j] == nil {
				return false
			}
			return sortedPlayers[i].PlayerName < sortedPlayers[j].PlayerName
		})

		for _, playerStats := range sortedPlayers {
			if playerStats == nil {
				continue
			}
			
			// Add map-specific stats
			for mapName, mapStats := range playerStats.MapStats {
				// Apply filters
				if st.filterMap != "" && mapName != st.filterMap {
					continue
				}

				// Add rows for T side and CT side separately
				for _, side := range []string{"T", "CT"} {
					if st.filterSide != "" && side != st.filterSide {
						continue
					}

					if sideStats, ok := mapStats.SideStats[side]; ok {
						st.addDataRow(row, playerStats.PlayerName, mapName, side, sideStats)
						row++
					}
				}
				
				// Add per-map summary row (T+CT combined) if not filtering by side
				if st.filterSide == "" {
					st.addMapSummaryRow(row, playerStats.PlayerName, mapName, mapStats)
					row++
				}
			}

			// Add overall row
			if st.filterMap == "" && st.filterSide == "" && playerStats.OverallStats != nil {
				st.addOverallRow(row, playerStats.PlayerName, playerStats.OverallStats)
				row++
			}
		}
	}
}

func (st *StatisticsTable) addDataRow(row int, playerName, mapName, side string,
	stats *SideStatistics) {
	if stats == nil {
		return
	}
	
	cols := []string{
		playerName,
		mapName,
		side,
		fmt.Sprintf("%.1f", stats.KAST),
		fmt.Sprintf("%.1f", stats.ADR),
		fmt.Sprintf("%.2f", stats.KD),
		fmt.Sprintf("%d", stats.Kills),
		fmt.Sprintf("%d", stats.Deaths),
		fmt.Sprintf("%d", stats.FirstKills),
		fmt.Sprintf("%d", stats.FirstDeaths),
		fmt.Sprintf("%d", stats.TradeKills),
		fmt.Sprintf("%d", stats.TradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorWhite)
		st.table.SetCell(row, col, cell)
	}
}

func (st *StatisticsTable) addMapSummaryRow(row int, playerName, mapName string, mapStats *MapStatistics) {
	// Defensive check
	if mapStats == nil || mapStats.SideStats == nil {
		return
	}
	
	// Calculate combined T+CT statistics for this map
	var totalKills, totalDeaths, totalAssists int
	var totalFirstKills, totalFirstDeaths int
	var totalTradeKills, totalTradeDeaths int
	var totalHeadshots, totalRoundsPlayed int
	var weightedKAST, weightedADR float64
	
	for _, sideStats := range mapStats.SideStats {
		if sideStats == nil {
			continue
		}
		totalKills += sideStats.Kills
		totalDeaths += sideStats.Deaths
		totalAssists += sideStats.Assists
		totalFirstKills += sideStats.FirstKills
		totalFirstDeaths += sideStats.FirstDeaths
		totalTradeKills += sideStats.TradeKills
		totalTradeDeaths += sideStats.TradeDeaths
		totalHeadshots += sideStats.Headshots
		totalRoundsPlayed += sideStats.RoundsPlayed
		
		// Weighted average for KAST and ADR
		weightedKAST += (sideStats.KAST / 100.0) * float64(sideStats.RoundsPlayed)
		weightedADR += sideStats.ADR * float64(sideStats.RoundsPlayed)
	}
	
	// Calculate averages
	kast := 0.0
	adr := 0.0
	if totalRoundsPlayed > 0 {
		kast = (weightedKAST / float64(totalRoundsPlayed)) * 100.0
		adr = weightedADR / float64(totalRoundsPlayed)
	}
	
	// Calculate K/D
	kd := 0.0
	if totalDeaths > 0 {
		kd = float64(totalKills) / float64(totalDeaths)
	} else if totalKills > 0 {
		kd = float64(totalKills)
	}
	
	cols := []string{
		playerName,
		mapName,
		"Both",
		fmt.Sprintf("%.1f", kast),
		fmt.Sprintf("%.1f", adr),
		fmt.Sprintf("%.2f", kd),
		fmt.Sprintf("%d", totalKills),
		fmt.Sprintf("%d", totalDeaths),
		fmt.Sprintf("%d", totalFirstKills),
		fmt.Sprintf("%d", totalFirstDeaths),
		fmt.Sprintf("%d", totalTradeKills),
		fmt.Sprintf("%d", totalTradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorAqua).
			SetAttributes(tcell.AttrBold)
		st.table.SetCell(row, col, cell)
	}
}

func (st *StatisticsTable) addOverallRow(row int, playerName string, stats *OverallStatistics) {
	if stats == nil {
		return
	}
	
	cols := []string{
		playerName,
		"Overall",
		"All",
		fmt.Sprintf("%.1f", stats.KAST),
		fmt.Sprintf("%.1f", stats.ADR),
		fmt.Sprintf("%.2f", stats.KD),
		fmt.Sprintf("%d", stats.Kills),
		fmt.Sprintf("%d", stats.Deaths),
		fmt.Sprintf("%d", stats.FirstKills),
		fmt.Sprintf("%d", stats.FirstDeaths),
		fmt.Sprintf("%d", stats.TradeKills),
		fmt.Sprintf("%d", stats.TradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorGreen).
			SetAttributes(tcell.AttrBold)
		st.table.SetCell(row, col, cell)
	}
}

func (st *StatisticsTable) SetFilter(mapFilter, sideFilter string) {
	st.filterMap = mapFilter
	st.filterSide = sideFilter
	st.renderTable()
}

// ============================================================================
// FORM CREATION AND VALIDATION
// ============================================================================

func createPlayerInputForm() *tview.Form {
	form := tview.NewForm()
	
	form.SetBorder(true)
	form.SetTitle("Player Configuration")
	form.SetTitleAlign(tview.AlignLeft)

	// Add 5 player input pairs
	for i := 1; i <= 5; i++ {
		playerLabel := fmt.Sprintf("Player %d Name", i)
		steamLabel := fmt.Sprintf("Player %d SteamID64", i)

		form.AddInputField(playerLabel, "", 30, nil, nil)
		form.AddInputField(steamLabel, "", 17, validateSteamID64, nil)
	}

	// Add base path input
	form.AddInputField("Demo Base Path", "", 50, nil, nil)

	// Add buttons
	form.AddButton("Analyze", nil) // Handler added later
	form.AddButton("Clear", nil)

	return form
}

// validateSteamID64 ensures only numeric input for SteamID64
func validateSteamID64(text string, lastChar rune) bool {
	// Allow empty string or only digits
	if text == "" {
		return true
	}
	// Check if character is a digit
	if lastChar < '0' || lastChar > '9' {
		return false
	}
	// Limit to 17 characters (SteamID64 length)
	return len(text) <= 17
}

// ============================================================================
// BUTTON HANDLERS
// ============================================================================

func (u *UI) setupFormHandlers(form *tview.Form) {
	// Get button indices (assuming Analyze=0, Clear=1)
	analyzeIdx := form.GetButtonCount() - 2
	clearIdx := form.GetButtonCount() - 1

	// Set Analyze button handler
	form.GetButton(analyzeIdx).SetSelectedFunc(func() {
		u.onAnalyzeClicked(form)
	})

	// Set Clear button handler
	form.GetButton(clearIdx).SetSelectedFunc(func() {
		u.onClearClicked(form)
	})
}

func (u *UI) onAnalyzeClicked(form *tview.Form) {
	// Collect form data
	config := u.extractConfigFromForm(form)

	// Validate base path first (required regardless of player count)
	if config.BasePath == "" {
		u.logEvent("Error: Demo base path must be specified")
		return
	}

	if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
		u.logEvent(fmt.Sprintf("Error: Path does not exist: %s", config.BasePath))
		return
	}

	// Validate at least one player is specified
	validPlayers := 0
	for _, player := range config.Players {
		if player.SteamID64 != "" {
			validPlayers++
		}
	}

	if validPlayers == 0 {
		u.logEvent("Error: At least one player with SteamID64 must be specified")
		return
	}

	// Start analysis (in goroutine to keep UI responsive)
	go u.runAnalysis(config)
}

func (u *UI) onClearClicked(form *tview.Form) {
	// Reset all form fields
	formItemCount := form.GetFormItemCount()
	for i := 0; i < formItemCount; i++ {
		if field, ok := form.GetFormItem(i).(*tview.InputField); ok {
			field.SetText("")
		}
	}
	u.logEvent("Form cleared")
}

func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig {
	config := AnalysisConfig{}

	// Extract player data (5 pairs of name + steamID)
	for i := 0; i < 5; i++ {
		nameIdx := i * 2
		steamIdx := i*2 + 1

		if nameField, ok := form.GetFormItem(nameIdx).(*tview.InputField); ok {
			config.Players[i].Name = nameField.GetText()
		}
		if steamField, ok := form.GetFormItem(steamIdx).(*tview.InputField); ok {
			config.Players[i].SteamID64 = steamField.GetText()
		}
	}

	// Extract base path (index 10 = after 5 player pairs)
	if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
		config.BasePath = pathField.GetText()
	}

	return config
}

// ============================================================================
// ANALYSIS EXECUTION
// ============================================================================

func (u *UI) runAnalysis(config AnalysisConfig) {
	// Add panic recovery to catch crashes and log them
	defer func() {
		if r := recover(); r != nil {
			u.logEvent(fmt.Sprintf("PANIC during analysis: %v", r))
		}
	}()
	
	u.logEvent("Starting analysis...")

	// Extract valid SteamIDs
	var steamIDs []string
	for _, player := range config.Players {
		if player.SteamID64 != "" {
			steamIDs = append(steamIDs, player.SteamID64)
			u.logEvent(fmt.Sprintf("Tracking player: %s (%s)",
				player.Name, player.SteamID64))
		}
	}

	// Gather demos
	u.logEvent(fmt.Sprintf("Searching for demos in: %s", config.BasePath))
	matches, err := GatherAllDemosFromPath(config.BasePath)

	if err != nil {
		u.logEvent(fmt.Sprintf("Warning during demo gathering: %v", err))
	}

	if len(matches) == 0 {
		u.logEvent("Error: No demo files found or all demos failed to parse")
		return
	}

	u.logEvent(fmt.Sprintf("Found %d demos, starting analysis...", len(matches)))

	// Process matches
	result, err := ProcessMatches(matches, steamIDs)
	if err != nil {
		u.logEvent(fmt.Sprintf("Error during analysis: %v", err))
		return
	}

	// Display results
	u.logEvent(fmt.Sprintf("Analysis complete! Processed %d matches", result.TotalMatches))
	u.logEvent(fmt.Sprintf("Found stats for %d players across %d maps",
		len(result.PlayerStats), len(result.MapList)))

	u.QueueUpdate(func() {
		u.statsTable.UpdateData(result)
	})
}

// ============================================================================
// UI INITIALIZATION
// ============================================================================

func New() *UI {
	app := tview.NewApplication()

	// Create components
	form := createPlayerInputForm()
	eventLog := newEventLog(50) // Keep last 50 events
	statsTable := newStatisticsTable()

	// Create layout
	leftPanel := form

	middlePanel := eventLog.textView
	middlePanel.SetBorder(true).
		SetTitle("Event Log").
		SetTitleAlign(tview.AlignLeft)

	bottomPanel := statsTable.table

	// Assemble layout with proper sizing
	rightColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(middlePanel, eventLogHeight, 0, false). // Fixed height for event log
		AddItem(bottomPanel, 0, 1, false)               // Rest for statistics table

	mainLayout := tview.NewFlex().
		AddItem(leftPanel, 0, 1, true).     // Left gets 1/3
		AddItem(rightColumn, 0, 2, false)   // Right gets 2/3

	pages := tview.NewPages().AddPage("main", mainLayout, true, true)

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC, tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
	})

	ui := &UI{
		App:        app,
		Pages:      pages,
		Root:       mainLayout,
		form:       form,
		eventLog:   eventLog,
		statsTable: statsTable,
	}

	// Setup handlers after UI is created
	ui.setupFormHandlers(form)

	return ui
}

// ============================================================================
// PUBLIC METHODS
// ============================================================================

func (u *UI) Start() error {
	return u.App.Run()
}

func (u *UI) Stop() {
	u.App.Stop()
}

func (u *UI) QueueUpdate(fn func()) {
	u.App.QueueUpdateDraw(fn)
}

func (u *UI) logEvent(message string) {
	u.QueueUpdate(func() {
		u.eventLog.Log(message)
	})
}
