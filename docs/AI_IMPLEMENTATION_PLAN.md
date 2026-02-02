# Manalyzer Implementation Plan - AI Agent Ready

**Date:** 2026-02-02  
**Status:** READY FOR IMPLEMENTATION  
**Estimated Effort:** 22-30 hours over 3-4 weeks

---

## Overview

Implement 4 major improvements to the Manalyzer CS:GO demo analyzer:

1. **Refactoring** - Split large files into focused modules
2. **Persistent Storage** - Save player configurations between sessions
3. **Filtering & Sorting** - Interactive table controls
4. **Data Visualization** - Web dashboard with interactive charts

**Approach:** Hybrid visualization (keep terminal UI, add optional web dashboard)

---

## Technical Stack

**Current:**
- Go 1.23+
- tview (terminal UI)
- cs-demo-analyzer (demo parsing)

**Adding:**
- `github.com/go-echarts/go-echarts/v2` (visualization only)
- Standard library `encoding/json` (config persistence)

---

## Implementation Phases

### Phase 1: Refactoring (2 days, ~6 hours)

Split `src/gui.go` (605 lines) into focused files.

**New Files:**
- `src/gui_eventlog.go` - EventLog component
- `src/gui_form.go` - Form and handlers  
- `src/gui_table.go` - StatisticsTable with sorting
- `src/config.go` - Configuration structures

**Critical Changes:**
- Add `config *Config` field to UI struct
- Add `app *tview.Application`, `sortColumn int`, `sortDesc bool` fields to StatisticsTable
- Update function signatures for config support

### Phase 2: Persistent Storage (2 days, ~4 hours)

Auto-save/load player configuration to JSON file.

**Location:** `~/.config/manalyzer/config.json` (or OS equivalent)

**Features:**
- Auto-load on startup
- Auto-save on analyze
- "Save Config" button for explicit saves

### Phase 3: Filtering & Sorting (1 week, ~8 hours)

Add interactive table controls.

**Features:**
- Click column headers to sort (with ▲/▼ indicators)
- Filter dropdowns for Map and Side (T/CT/All)
- Combined filtering and sorting

### Phase 4: Visualization (2 weeks, ~8 hours)

Web dashboard with go-echarts charts.

**Features:**
- "Visualize" button opens browser
- HTTP server on localhost:8080-8090 (auto-select available port)
- 3 chart types: Player comparison, T vs CT performance, Map breakdown
- Browser auto-open with fallbacks

---

## Complete Code Examples

### Phase 1.1: EventLog Component

**Create:** `src/gui_eventlog.go`

```go
package manalyzer

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// EventLog displays timestamped event messages.
type EventLog struct {
	textView *tview.TextView
	maxLines int
	lines    []string
}

func newEventLog(maxLines int) *EventLog {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	
	tv.SetBorder(true)
	tv.SetTitle("Event Log")

	tv.SetChangedFunc(func() {
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

	if len(el.lines) > el.maxLines {
		el.lines = el.lines[len(el.lines)-el.maxLines:]
	}

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

	var builder strings.Builder
	for i, l := range el.lines {
		builder.WriteString(l)
		if i < len(el.lines)-1 {
			builder.WriteString("\n")
		}
	}
	el.textView.SetText(builder.String())
}
```

---

### Phase 1.2: StatisticsTable Component

**Create:** `src/gui_table.go`

```go
package manalyzer

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatisticsTable displays player statistics.
type StatisticsTable struct {
	table      *tview.Table
	data       *WrangleResult
	filterMap  string
	filterSide string
	app        *tview.Application  // NEW: For UI updates
	sortColumn int                 // NEW: Current sort column (0-11)
	sortDesc   bool                // NEW: Sort direction
}

func newStatisticsTable(app *tview.Application) *StatisticsTable {
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0).
		SetSelectable(true, false)
	
	table.SetBorder(true)
	table.SetTitle("Player Statistics")

	return &StatisticsTable{
		table:      table,
		app:        app,
		filterMap:  "",
		filterSide: "",
		sortColumn: 0,
		sortDesc:   false,
	}
}

func (st *StatisticsTable) UpdateData(result *WrangleResult) {
	st.data = result
	st.renderTable()
}

func (st *StatisticsTable) SetFilter(mapFilter, sideFilter string) {
	st.filterMap = mapFilter
	st.filterSide = sideFilter
	st.renderTable()
}

func (st *StatisticsTable) toggleSort(columnIndex int) {
	if st.sortColumn == columnIndex {
		st.sortDesc = !st.sortDesc
	} else {
		st.sortColumn = columnIndex
		if columnIndex == 0 || columnIndex == 1 {
			st.sortDesc = false  // Ascending for names
		} else {
			st.sortDesc = true   // Descending for stats
		}
	}
}

func (st *StatisticsTable) sortData(players []*PlayerStats) []*PlayerStats {
	sorted := make([]*PlayerStats, len(players))
	copy(sorted, players)
	
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i] == nil || sorted[j] == nil {
			return false
		}
		
		var less bool
		switch st.sortColumn {
		case 0: // Player name
			less = sorted[i].PlayerName < sorted[j].PlayerName
		case 3: // KAST%
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.KAST < sorted[j].OverallStats.KAST
			}
		case 4: // ADR
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.ADR < sorted[j].OverallStats.ADR
			}
		case 5: // K/D
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.KD < sorted[j].OverallStats.KD
			}
		case 6: // Kills
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.Kills < sorted[j].OverallStats.Kills
			}
		case 7: // Deaths
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.Deaths < sorted[j].OverallStats.Deaths
			}
		case 8: // FK
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.FirstKills < sorted[j].OverallStats.FirstKills
			}
		case 9: // FD
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.FirstDeaths < sorted[j].OverallStats.FirstDeaths
			}
		case 10: // TK
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.TradeKills < sorted[j].OverallStats.TradeKills
			}
		case 11: // TD
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.TradeDeaths < sorted[j].OverallStats.TradeDeaths
			}
		default:
			less = sorted[i].PlayerName < sorted[j].PlayerName
		}
		
		if st.sortDesc {
			return !less
		}
		return less
	})
	
	return sorted
}

func (st *StatisticsTable) renderTable() {
	st.table.Clear()

	// Header row
	headers := []string{"Player", "Map", "Side", "KAST%", "ADR", "K/D",
		"Kills", "Deaths", "FK", "FD", "TK", "TD"}

	for col, header := range headers {
		columnIndex := col  // Capture loop variable
		
		displayHeader := header
		if st.sortColumn == columnIndex {
			if st.sortDesc {
				displayHeader += " ▼"
			} else {
				displayHeader += " ▲"
			}
		}
		
		cell := tview.NewTableCell(displayHeader).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetAttributes(tcell.AttrBold)
		
		// Make stat columns clickable
		if columnIndex >= 3 {
			cell.SetClickedFunc(func() bool {
				st.toggleSort(columnIndex)
				st.app.QueueUpdateDraw(func() {
					st.renderTable()
				})
				return true
			})
		}
		
		st.table.SetCell(0, col, cell)
	}

	// Data rows
	row := 1
	if st.data != nil {
		sortedPlayers := make([]*PlayerStats, 0, len(st.data.PlayerStats))
		
		for _, ps := range st.data.PlayerStats {
			if ps != nil {
				sortedPlayers = append(sortedPlayers, ps)
			}
		}
		
		sortedPlayers = st.sortData(sortedPlayers)

		for _, playerStats := range sortedPlayers {
			if playerStats == nil {
				continue
			}
			
			for mapName, mapStats := range playerStats.MapStats {
				if st.filterMap != "" && mapName != st.filterMap {
					continue
				}

				for _, side := range []string{"T", "CT"} {
					if st.filterSide != "" && side != st.filterSide {
						continue
					}

					if sideStats, ok := mapStats.SideStats[side]; ok {
						st.addDataRow(row, playerStats.PlayerName, mapName, side, sideStats)
						row++
					}
				}
				
				if st.filterSide == "" {
					st.addMapSummaryRow(row, playerStats.PlayerName, mapName, mapStats)
					row++
				}
			}

			if st.filterMap == "" && st.filterSide == "" && playerStats.OverallStats != nil {
				st.addOverallRow(row, playerStats.PlayerName, playerStats.OverallStats)
				row++
			}
		}
	}
}

func (st *StatisticsTable) addDataRow(row int, playerName, mapName, side string, stats *SideStatistics) {
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
	if mapStats == nil || mapStats.SideStats == nil {
		return
	}
	
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
		
		weightedKAST += (sideStats.KAST / 100.0) * float64(sideStats.RoundsPlayed)
		weightedADR += sideStats.ADR * float64(sideStats.RoundsPlayed)
	}
	
	kast := 0.0
	adr := 0.0
	if totalRoundsPlayed > 0 {
		kast = (weightedKAST / float64(totalRoundsPlayed)) * 100.0
		adr = weightedADR / float64(totalRoundsPlayed)
	}
	
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
```

---

### Phase 1.3: Form Component

**Create:** `src/gui_form.go`

```go
package manalyzer

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
)

// PlayerInput represents user input for player tracking.
type PlayerInput struct {
	Name      string
	SteamID64 string
}

// AnalysisConfig holds configuration for analysis.
type AnalysisConfig struct {
	Players  [5]PlayerInput
	BasePath string
}

func createPlayerInputForm() *tview.Form {
	return createPlayerInputFormWithConfig(DefaultConfig())
}

func createPlayerInputFormWithConfig(config *Config) *tview.Form {
	form := tview.NewForm()
	
	form.SetBorder(true)
	form.SetTitle("Player Configuration")
	form.SetTitleAlign(tview.AlignLeft)

	for i := 0; i < 5; i++ {
		playerName := ""
		steamID := ""
		if i < len(config.Players) {
			playerName = config.Players[i].Name
			steamID = config.Players[i].SteamID64
		}
		
		playerLabel := fmt.Sprintf("Player %d Name", i+1)
		steamLabel := fmt.Sprintf("Player %d SteamID64", i+1)

		form.AddInputField(playerLabel, playerName, 30, nil, nil)
		form.AddInputField(steamLabel, steamID, 17, validateSteamID64, nil)
	}

	form.AddInputField("Demo Base Path", config.LastDemoPath, 50, nil, nil)

	form.AddButton("Analyze", nil)
	form.AddButton("Clear", nil)
	form.AddButton("Save Config", nil)
	form.AddButton("Visualize", nil)

	return form
}

func validateSteamID64(text string, lastChar rune) bool {
	if text == "" {
		return true
	}
	if lastChar < '0' || lastChar > '9' {
		return false
	}
	return len(text) <= 17
}

func (u *UI) setupFormHandlers(form *tview.Form) {
	analyzeIdx := form.GetButtonCount() - 4
	clearIdx := form.GetButtonCount() - 3
	saveIdx := form.GetButtonCount() - 2
	visualizeIdx := form.GetButtonCount() - 1

	form.GetButton(analyzeIdx).SetSelectedFunc(func() {
		u.onAnalyzeClicked(form)
	})

	form.GetButton(clearIdx).SetSelectedFunc(func() {
		u.onClearClicked(form)
	})

	form.GetButton(saveIdx).SetSelectedFunc(func() {
		u.onSaveConfigClicked(form)
	})

	form.GetButton(visualizeIdx).SetSelectedFunc(func() {
		u.onVisualizeClicked()
	})
}

func (u *UI) onAnalyzeClicked(form *tview.Form) {
	config := u.extractConfigFromForm(form)

	if config.BasePath == "" {
		u.logEvent("Error: Demo base path must be specified")
		return
	}

	if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
		u.logEvent(fmt.Sprintf("Error: Path does not exist: %s", config.BasePath))
		return
	}

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

	// Save config if auto-save enabled
	if u.config.Preferences.AutoSave {
		persistConfig := u.buildConfigFromForm(form)
		if err := SaveConfig(persistConfig); err != nil {
			u.logEvent(fmt.Sprintf("Warning: Failed to save config: %v", err))
		} else {
			u.config = persistConfig
		}
	}

	go u.runAnalysis(config)
}

func (u *UI) onClearClicked(form *tview.Form) {
	formItemCount := form.GetFormItemCount()
	for i := 0; i < formItemCount; i++ {
		if field, ok := form.GetFormItem(i).(*tview.InputField); ok {
			field.SetText("")
		}
	}
	u.logEvent("Form cleared")
}

func (u *UI) onSaveConfigClicked(form *tview.Form) {
	config := u.buildConfigFromForm(form)
	if err := SaveConfig(config); err != nil {
		u.logEvent(fmt.Sprintf("Error saving config: %v", err))
	} else {
		u.config = config
		u.logEvent("Configuration saved successfully")
	}
}

func (u *UI) onVisualizeClicked() {
	if u.statsTable.data == nil {
		u.logEvent("Error: No data to visualize. Run analysis first.")
		return
	}
	
	u.logEvent("Starting visualization server...")
	
	url, err := StartVisualizationServer(u.statsTable.data)
	if err != nil {
		u.logEvent(fmt.Sprintf("Error: Failed to start server: %v", err))
		return
	}
	
	u.logEvent(fmt.Sprintf("Visualization server started at %s", url))
	
	time.Sleep(500 * time.Millisecond)
	
	if err := openBrowser(url); err != nil {
		u.logEvent(fmt.Sprintf("Could not open browser: %v", err))
		u.logEvent(fmt.Sprintf("Visit %s manually to view charts", url))
	} else {
		u.logEvent("Visualization dashboard opened in browser")
	}
}

func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig {
	config := AnalysisConfig{}

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

	if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
		config.BasePath = pathField.GetText()
	}

	return config
}

func (u *UI) buildConfigFromForm(form *tview.Form) *Config {
	config := &Config{
		Version: 1,
		Players: make([]PlayerConfig, 0),
		Preferences: Preferences{
			AutoSave: true,
		},
	}
	
	for i := 0; i < 5; i++ {
		nameIdx := i * 2
		steamIdx := i*2 + 1

		var name, steamID string
		if nameField, ok := form.GetFormItem(nameIdx).(*tview.InputField); ok {
			name = nameField.GetText()
		}
		if steamField, ok := form.GetFormItem(steamIdx).(*tview.InputField); ok {
			steamID = steamField.GetText()
		}
		
		if steamID != "" {
			config.Players = append(config.Players, PlayerConfig{
				Name:      name,
				SteamID64: steamID,
			})
		}
	}

	if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
		config.LastDemoPath = pathField.GetText()
	}

	return config
}

func (u *UI) runAnalysis(config AnalysisConfig) {
	defer func() {
		if r := recover(); r != nil {
			u.logEvent(fmt.Sprintf("PANIC during analysis: %v", r))
		}
	}()
	
	u.logEvent("Starting analysis...")

	var steamIDs []string
	for _, player := range config.Players {
		if player.SteamID64 != "" {
			steamIDs = append(steamIDs, player.SteamID64)
			u.logEvent(fmt.Sprintf("Tracking player: %s (%s)", player.Name, player.SteamID64))
		}
	}

	u.logEvent(fmt.Sprintf("Searching for demos in: %s", config.BasePath))
	matches, err := GatherAllDemosFromPath(config.BasePath)

	if err != nil {
		if len(matches) == 0 {
			u.logEvent(fmt.Sprintf("Error: %v", err))
			return
		}
		u.logEvent(fmt.Sprintf("Warning during demo gathering: %v", err))
	}

	if len(matches) == 0 {
		u.logEvent("Error: No demo files found or all demos failed to parse")
		return
	}

	u.logEvent(fmt.Sprintf("Found %d demos, starting analysis...", len(matches)))

	result, err := ProcessMatches(matches, steamIDs)
	if err != nil {
		u.logEvent(fmt.Sprintf("Error during analysis: %v", err))
		return
	}

	u.logEvent(fmt.Sprintf("Analysis complete! Processed %d matches", result.TotalMatches))
	u.logEvent(fmt.Sprintf("Found stats for %d players across %d maps",
		len(result.PlayerStats), len(result.MapList)))

	u.QueueUpdate(func() {
		u.statsTable.UpdateData(result)
	})
}
```

---

### Phase 1.4: Config Structures

**Create:** `src/config.go`

```go
package manalyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents persistent application configuration.
type Config struct {
	Version      int            `json:"version"`
	Players      []PlayerConfig `json:"players"`
	LastDemoPath string         `json:"lastDemoPath,omitempty"`
	Preferences  Preferences    `json:"preferences"`
}

// PlayerConfig stores a player's name and SteamID64.
type PlayerConfig struct {
	Name      string `json:"name"`
	SteamID64 string `json:"steamID64"`
}

// Preferences stores UI preferences.
type Preferences struct {
	AutoSave bool `json:"autoSave"`
}

// DefaultConfig returns a config with defaults.
func DefaultConfig() *Config {
	return &Config{
		Version:     1,
		Players:     make([]PlayerConfig, 0),
		Preferences: Preferences{AutoSave: true},
	}
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".manalyzer", "config.json"), nil
	}
	return filepath.Join(configDir, "manalyzer", "config.json"), nil
}

// LoadConfig loads configuration from disk.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), nil
	}

	return &config, nil
}

// SaveConfig saves configuration to disk.
func SaveConfig(config *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
```

---

### Phase 1.5: Update Core UI

**Update:** `src/gui.go` - Keep only core functions

```go
package manalyzer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	eventLogHeight = 5
)

// UI manages the terminal user interface.
type UI struct {
	App        *tview.Application
	Pages      *tview.Pages
	Root       *tview.Flex
	form       *tview.Form
	eventLog   *EventLog
	statsTable *StatisticsTable
	config     *Config  // NEW: Store loaded configuration
}

func New() *UI {
	app := tview.NewApplication()

	// Load config
	config, err := LoadConfig()
	if err != nil {
		config = DefaultConfig()
	}

	// Create components
	form := createPlayerInputFormWithConfig(config)
	eventLog := newEventLog(50)
	statsTable := newStatisticsTable(app)  // Pass app reference

	// Create layout
	leftPanel := form

	middlePanel := eventLog.textView
	middlePanel.SetBorder(true).
		SetTitle("Event Log").
		SetTitleAlign(tview.AlignLeft)

	bottomPanel := statsTable.table

	rightColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(middlePanel, eventLogHeight, 0, false).
		AddItem(bottomPanel, 0, 1, false)

	mainLayout := tview.NewFlex().
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightColumn, 0, 2, false)

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
		config:     config,
	}

	ui.setupFormHandlers(form)

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

func (u *UI) logEvent(message string) {
	u.QueueUpdate(func() {
		u.eventLog.Log(message)
	})
}
```

---

### Phase 4: Visualization Server

**Create:** `src/visualize.go`

```go
package manalyzer

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// StartVisualizationServer starts HTTP server for visualization.
func StartVisualizationServer(data *WrangleResult) (string, error) {
	// Find available port
	var listener net.Listener
	var port int
	var err error
	
	for port = 8080; port <= 8090; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return "", fmt.Errorf("no available ports in range 8080-8090")
	}
	
	// Set up handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler(data))
	mux.HandleFunc("/player-comparison", playerComparisonHandler(data))
	mux.HandleFunc("/side-performance", sidePerformanceHandler(data))
	mux.HandleFunc("/map-breakdown", mapBreakdownHandler(data))
	
	// Start server
	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	
	url := fmt.Sprintf("http://localhost:%d", port)
	return url, nil
}

// openBrowser opens URL in system browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "linux":
		browsers := []string{"xdg-open", "sensible-browser", "firefox", "chromium-browser", "google-chrome"}
		for _, browser := range browsers {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, url)
				return cmd.Start()
			}
		}
		return fmt.Errorf("no browser found")
		
	case "darwin":
		cmd = exec.Command("open", url)
		return cmd.Start()
		
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		return cmd.Start()
		
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// dashboardHandler serves the main dashboard page.
func dashboardHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Manalyzer - Player Statistics Dashboard</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background: #1a1a1a; color: #fff; }
		h1 { text-align: center; color: #4CAF50; }
		.chart-container { margin: 20px 0; }
		a { color: #4CAF50; text-decoration: none; margin: 0 15px; }
		a:hover { text-decoration: underline; }
		.nav { text-align: center; margin: 20px 0; padding: 15px; background: #2a2a2a; }
	</style>
</head>
<body>
	<h1>🎮 CS:GO Player Statistics Dashboard</h1>
	<div class="nav">
		<a href="/player-comparison">Player Comparison</a>
		<a href="/side-performance">T vs CT Performance</a>
		<a href="/map-breakdown">Map Breakdown</a>
	</div>
	<div class="chart-container">
		<h2>Quick Overview</h2>
		<p>Select a chart from the navigation above to view detailed statistics.</p>
		<p><strong>Total Matches Analyzed:</strong> ` + fmt.Sprintf("%d", data.TotalMatches) + `</p>
		<p><strong>Players Tracked:</strong> ` + fmt.Sprintf("%d", len(data.PlayerStats)) + `</p>
		<p><strong>Maps:</strong> ` + strings.Join(data.MapList, ", ") + `</p>
	</div>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

// playerComparisonHandler creates player comparison bar chart.
func playerComparisonHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "Player Comparison",
				Subtitle: "Overall Performance Metrics",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: true}),
			charts.WithLegendOpts(opts.Legend{Show: true}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		var players []string
		kastData := make([]opts.BarData, 0)
		adrData := make([]opts.BarData, 0)
		kdData := make([]opts.BarData, 0)
		
		for _, player := range data.PlayerStats {
			if player != nil && player.OverallStats != nil {
				players = append(players, player.PlayerName)
				kastData = append(kastData, opts.BarData{Value: player.OverallStats.KAST})
				adrData = append(adrData, opts.BarData{Value: player.OverallStats.ADR})
				kdData = append(kdData, opts.BarData{Value: player.OverallStats.KD * 100})
			}
		}
		
		bar.SetXAxis(players)
		bar.AddSeries("KAST%", kastData)
		bar.AddSeries("ADR", adrData)
		bar.AddSeries("K/D x100", kdData)
		
		bar.Render(w)
	}
}

// sidePerformanceHandler creates T vs CT comparison chart.
func sidePerformanceHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "T vs CT Performance",
				Subtitle: "Side-Specific Statistics",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: true}),
			charts.WithLegendOpts(opts.Legend{Show: true}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		var players []string
		tKAST := make([]opts.BarData, 0)
		ctKAST := make([]opts.BarData, 0)
		
		for _, player := range data.PlayerStats {
			if player != nil {
				players = append(players, player.PlayerName)
				
				var tKastSum, ctKastSum float64
				var tRounds, ctRounds int
				
				for _, mapStat := range player.MapStats {
					if tStats, ok := mapStat.SideStats["T"]; ok {
						tKastSum += (tStats.KAST / 100.0) * float64(tStats.RoundsPlayed)
						tRounds += tStats.RoundsPlayed
					}
					if ctStats, ok := mapStat.SideStats["CT"]; ok {
						ctKastSum += (ctStats.KAST / 100.0) * float64(ctStats.RoundsPlayed)
						ctRounds += ctStats.RoundsPlayed
					}
				}
				
				tAvg := 0.0
				if tRounds > 0 {
					tAvg = (tKastSum / float64(tRounds)) * 100.0
				}
				ctAvg := 0.0
				if ctRounds > 0 {
					ctAvg = (ctKastSum / float64(ctRounds)) * 100.0
				}
				
				tKAST = append(tKAST, opts.BarData{Value: tAvg})
				ctKAST = append(ctKAST, opts.BarData{Value: ctAvg})
			}
		}
		
		bar.SetXAxis(players)
		bar.AddSeries("T Side KAST%", tKAST)
		bar.AddSeries("CT Side KAST%", ctKAST)
		
		bar.Render(w)
	}
}

// mapBreakdownHandler creates map performance heatmap.
func mapBreakdownHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hm := charts.NewHeatMap()
		hm.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "Map Performance Breakdown",
				Subtitle: "KAST% by Player and Map",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: true}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		mapSet := make(map[string]bool)
		for _, mapName := range data.MapList {
			mapSet[mapName] = true
		}
		maps := make([]string, 0, len(mapSet))
		for m := range mapSet {
			maps = append(maps, m)
		}
		
		var players []string
		for _, player := range data.PlayerStats {
			if player != nil {
				players = append(players, player.PlayerName)
			}
		}
		
		hmData := make([]opts.HeatMapData, 0)
		for pi, player := range data.PlayerStats {
			if player == nil {
				continue
			}
			for mi, mapName := range maps {
				kast := 0.0
				if mapStat, ok := player.MapStats[mapName]; ok {
					var kastSum float64
					var rounds int
					for _, sideStat := range mapStat.SideStats {
						kastSum += (sideStat.KAST / 100.0) * float64(sideStat.RoundsPlayed)
						rounds += sideStat.RoundsPlayed
					}
					if rounds > 0 {
						kast = (kastSum / float64(rounds)) * 100.0
					}
				}
				hmData = append(hmData, opts.HeatMapData{
					Value: [3]interface{}{mi, pi, kast},
				})
			}
		}
		
		hm.SetXAxis(maps)
		hm.SetYAxis(players)
		hm.AddSeries("KAST%", hmData)
		
		hm.Render(w)
	}
}
```

---

## Step-by-Step Implementation

### Phase 1: Refactoring

1. **Create gui_eventlog.go**
   - Copy EventLog struct and methods from gui.go
   - Remove from gui.go
   - Run: `go build`
   - Test: App should work identically

2. **Create gui_table.go**
   - Copy StatisticsTable with NEW fields (app, sortColumn, sortDesc)
   - Add toggleSort() and sortData() methods
   - Update renderTable() with sort indicators and click handlers
   - Update newStatisticsTable to accept app parameter
   - Remove from gui.go
   - Update New() in gui.go to pass app to newStatisticsTable
   - Run: `go build`
   - Test: Sorting should work when clicking headers

3. **Create gui_form.go**
   - Move all form-related functions
   - Add onVisualizeClicked() (calls will fail until Phase 4, that's OK)
   - Add buildConfigFromForm()
   - Remove from gui.go
   - Run: `go build`
   - Test: Form should work

4. **Create config.go**
   - Add all config structures and functions
   - Run: `go build`

5. **Update gui.go**
   - Add `config *Config` field to UI struct
   - Update New() to load config and pass to components
   - Keep only: New(), Start(), Stop(), QueueUpdate(), logEvent()
   - Run: `go build`
   - Test: Config should load on startup, form should be pre-filled

**Commit:** "refactor: split gui.go into focused modules"

---

### Phase 2: Persistent Storage

Config loading is already done in Phase 1. Just verify:

1. Delete any existing config file
2. Run app - should use defaults (empty form)
3. Enter player data
4. Click Analyze (with valid demo path)
5. Check: `~/.config/manalyzer/config.json` created
6. Restart app
7. Verify: Form is pre-filled

**Commit:** "feat: add persistent configuration"

---

### Phase 3: Filtering & Sorting

Sorting is already done in Phase 1. No additional code needed - just test:

1. Run app with analysis data
2. Click KAST% header - should sort descending with ▼
3. Click again - should sort ascending with ▲
4. Try other columns

**Commit:** "feat: add interactive table sorting"

---

### Phase 4: Visualization

1. **Add dependency:**
   ```bash
   go get github.com/go-echarts/go-echarts/v2
   ```

2. **Create visualize.go**
   - Copy complete code above
   - Run: `go build`

3. **Test:**
   - Run app with analysis data
   - Click "Visualize" button
   - Browser should open
   - Dashboard should display
   - All 3 chart links should work

**Commit:** "feat: add web visualization dashboard"

---

## Verification Checklist

### After Phase 1
- [ ] App compiles: `go build`
- [ ] App runs: `./manalyzer`
- [ ] UI looks identical to before
- [ ] Can enter data and analyze
- [ ] Results display in table
- [ ] Clicking stat headers sorts table
- [ ] Sort indicators (▲/▼) display

### After Phase 2
- [ ] Config file created on first analyze
- [ ] Config file location correct for your OS
- [ ] Restart app loads saved data
- [ ] "Save Config" button works explicitly
- [ ] Manual JSON edits are loaded

### After Phase 3
- [ ] All columns sortable (except Player, Map, Side)
- [ ] Sort direction toggles on repeated clicks
- [ ] Sort indicators move with column

### After Phase 4
- [ ] "Visualize" button appears
- [ ] Clicking starts server
- [ ] Browser opens automatically
- [ ] Dashboard page loads
- [ ] All 3 chart pages work
- [ ] Charts are interactive (hover, zoom)
- [ ] Works on your OS (Linux/Mac/Windows)

---

## Common Issues & Solutions

**Issue:** Compilation error "undefined: Config"
**Solution:** Make sure config.go is created and in same package

**Issue:** Sorting doesn't work
**Solution:** Verify app reference passed to newStatisticsTable

**Issue:** UI doesn't update when sorting
**Solution:** Verify click handler wraps renderTable in QueueUpdateDraw

**Issue:** Config not persisting
**Solution:** Check directory permissions, verify os.MkdirAll succeeds

**Issue:** Browser doesn't open
**Solution:** Check event log for manual URL, try different browser

**Issue:** Port 8080 in use
**Solution:** Code tries 8080-8090 automatically, check event log for actual port

---

## Success Criteria

- ✅ All 4 phases implemented
- ✅ App compiles without errors
- ✅ All existing functionality works
- ✅ Config persists between sessions
- ✅ Table is sortable and filterable
- ✅ Visualization opens in browser
- ✅ No data loss or corruption
- ✅ Cross-platform compatible

---

## Final Notes

**Critical Details:**
- Always use `columnIndex := col` to capture loop variable in closures
- Always wrap UI updates in `app.QueueUpdateDraw()`
- Config file permissions: 0600 (user-only)
- Port range: 8080-8090 (auto-select available)

**Time Estimates:**
- Phase 1: 6 hours
- Phase 2: 1 hour (mostly testing)
- Phase 3: 1 hour (mostly testing)
- Phase 4: 8 hours

**Total: 22-30 hours**

This plan is ready for implementation by an expert AI agent. Follow phases sequentially, test after each phase, commit incrementally.
