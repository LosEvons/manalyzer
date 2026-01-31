# Manalyzer Implementation Plan

## Overview
This document outlines the implementation plan for the Manalyzer application - a CS:GO demo analyzer that collects player statistics from demo files, filters them by specified players, and displays the results in a terminal UI.

## Architecture Overview

```
┌─────────────┐
│   gui.go    │ ← User inputs player info & demo path
└──────┬──────┘
       │ (base path)
       ↓
┌─────────────┐
│  gather.go  │ ← Finds and analyzes all demos
└──────┬──────┘
       │ ([]*api.Match)
       ↓
┌─────────────┐
│ wrangle.go  │ ← Filters & structures player data
└──────┬──────┘
       │ (PlayerStats)
       ↓
┌─────────────┐
│   gui.go    │ ← Displays tables in bottom panel
└─────────────┘
```

## 1. GUI Component (gui.go)

### 1.1 Input Form Structure
The left panel should be converted from a simple Box to a Form containing:

**Player Input Fields (5 sets):**
- Player 1 Name (InputField)
- Player 1 SteamID64 (InputField)
- Player 2 Name (InputField)
- Player 2 SteamID64 (InputField)
- Player 3 Name (InputField)
- Player 3 SteamID64 (InputField)
- Player 4 Name (InputField)
- Player 4 SteamID64 (InputField)
- Player 5 Name (InputField)
- Player 5 SteamID64 (InputField)

**Demo Base Path:**
- Base Path (InputField) - Path to folder containing demo files

**Action Buttons:**
- "Analyze" Button - Triggers the analysis pipeline
- "Clear" Button - Clears all input fields

### 1.2 Data Structures

```go
type PlayerInput struct {
    Name      string  // For display purposes only
    SteamID64 string  // Only SteamID64 format supported (17 digits)
}

type AnalysisConfig struct {
    Players  [5]PlayerInput  // Support 1-5 players (not all slots required)
    BasePath string
}
```

### 1.3 UI Layout Enhancement

**Current Layout:**
```
┌─────────┬──────────────┐
│  Left   │   Middle     │
│  Box    │   TextView   │
│         ├──────────────┤
│         │   Bottom Box │
└─────────┴──────────────┘
```

**New Layout:**
```
┌──────────────┬────────────────┐
│   Player     │  Event Log     │
│   Input      │  (3-5 rows)    │
│   Form       ├────────────────┤
│              │   Statistics   │
│   [Analyze]  │   Table(s)     │
│   [Clear]    │   (filterable) │
└──────────────┴────────────────┘
```

The middle panel should be compact (3-5 rows) to show recent events without taking too much space. It will display:
- Demo processing status
- Errors/warnings
- Analysis progress
- Completion messages

### 1.4 Implementation Steps

1. **Replace Left Box with Form**
   - Use `tview.NewForm()` instead of `tview.NewBox()`
   - Add 10 InputFields (5 name + 5 SteamID64)
   - Add 1 InputField for base path
   - Add 2 Buttons (Analyze, Clear)

2. **Add Form Validation**
   - SteamID64 fields should accept only numeric input (17 digits)
   - Base path should be validated for directory existence
   - At least one player pair (name + SteamID64) must be filled
   - Support analyzing 1-5 players (empty slots are ignored)

3. **Wire Button Handlers**
   - "Analyze" button → Collect form data → Call analysis pipeline
   - "Clear" button → Reset all form fields

4. **Update Middle Panel for Event Log**
   - Keep as TextView but limit height to 3-5 rows
   - Show scrolling log of events
   - Include timestamps for key events
   - Auto-scroll to bottom on new messages

5. **Update Bottom Panel for Tables**
   - Replace Bottom Box with `tview.Table` or multiple tables
   - Support filtering by map and/or side (T/CT)
   - Support pagination/scrolling for large datasets
   - Allow independent or combined filtering

### 1.5 User Interaction Flow

```
User fills form → Clicks "Analyze" 
  → GUI validates inputs
  → GUI creates AnalysisConfig
  → GUI calls gather.GatherAllDemosFromPath(basePath)
  → GUI calls wrangle.ProcessMatches(matches, steamIDs)
  → GUI displays results in bottom table
  → Status updates shown in middle TextView
```

### 1.6 Detailed UI Implementation Examples

#### 1.6.1 Creating the Input Form

```go
// In gui.go

func createPlayerInputForm() *tview.Form {
    form := tview.NewForm().
        SetBorder(true).
        SetTitle("Player Configuration").
        SetTitleAlign(tview.AlignLeft)
    
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
```

#### 1.6.2 Wiring Button Handlers

```go
// In gui.go

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
    
    // Validate base path
    if config.BasePath == "" {
        u.logEvent("Error: Demo base path must be specified")
        return
    }
    
    if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
        u.logEvent(fmt.Sprintf("Error: Path does not exist: %s", config.BasePath))
        return
    }
    
    // Start analysis (in goroutine to keep UI responsive)
    go u.runAnalysis(config)
}

func (u *UI) onClearClicked(form *tview.Form) {
    // Reset all form fields
    for i := 0; i < form.GetFormItemCount()-1; i++ { // -1 to exclude buttons
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
```

#### 1.6.3 Event Logging System

```go
// In gui.go

type EventLog struct {
    textView *tview.TextView
    maxLines int
    lines    []string
}

func newEventLog(maxLines int) *EventLog {
    tv := tview.NewTextView().
        SetDynamicColors(true).
        SetScrollable(true).
        SetBorder(true).
        SetTitle("Event Log")
    
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
    
    // Update display
    el.textView.Clear()
    for _, l := range el.lines {
        fmt.Fprintln(el.textView, l)
    }
}

func (el *EventLog) LogError(message string) {
    timestamp := time.Now().Format("15:04:05")
    line := fmt.Sprintf("[yellow]%s[-] [red]ERROR:[-] %s", timestamp, message)
    
    el.lines = append(el.lines, line)
    if len(el.lines) > el.maxLines {
        el.lines = el.lines[len(el.lines)-el.maxLines:]
    }
    
    el.textView.Clear()
    for _, l := range el.lines {
        fmt.Fprintln(el.textView, l)
    }
}

// Update UI struct to include EventLog
func (u *UI) logEvent(message string) {
    u.QueueUpdate(func() {
        u.eventLog.Log(message)
    })
}
```

#### 1.6.4 Statistics Table with Filtering

```go
// In gui.go

type StatisticsTable struct {
    table       *tview.Table
    data        *WrangleResult
    filterMap   string  // "" = all maps, or specific map name
    filterSide  string  // "" = all sides, "T", or "CT"
}

func newStatisticsTable() *StatisticsTable {
    table := tview.NewTable().
        SetBorders(true).
        SetFixed(1, 0). // Fix header row
        SetSelectable(true, false).
        SetBorder(true).
        SetTitle("Player Statistics (Press F to filter)")
    
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
        sortedPlayers := make([]*PlayerStats, len(st.data.PlayerStats))
        copy(sortedPlayers, st.data.PlayerStats)
        sort.Slice(sortedPlayers, func(i, j int) bool {
            return sortedPlayers[i].PlayerName < sortedPlayers[j].PlayerName
        })
        
        for _, playerStats := range sortedPlayers {
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
                    
                    st.addDataRow(row, playerStats.PlayerName, mapName, side, 
                                  mapStats.SideStats[side])
                    row++
                }
            }
            
            // Add overall row
            if st.filterMap == "" && st.filterSide == "" {
                st.addOverallRow(row, playerStats.PlayerName, playerStats.OverallStats)
                row++
            }
        }
    }
}

func (st *StatisticsTable) addDataRow(row int, playerName, mapName, side string, 
                                     stats *SideStatistics) {
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

func (st *StatisticsTable) SetFilter(mapFilter, sideFilter string) {
    st.filterMap = mapFilter
    st.filterSide = sideFilter
    st.renderTable()
}
```

#### 1.6.5 Complete UI Assembly

```go
// In gui.go

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
        AddItem(middlePanel, 5, 0, false).  // Fixed 5 rows for event log
        AddItem(bottomPanel, 0, 1, false)    // Rest for statistics table
    
    mainLayout := tview.NewFlex().
        AddItem(leftPanel, 0, 1, true).      // Left gets 1/3
        AddItem(rightColumn, 0, 2, false)    // Right gets 2/3
    
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
```

#### 1.6.6 Running Analysis with Progress Updates

```go
// In gui.go

func (u *UI) runAnalysis(config AnalysisConfig) {
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
    matches, err := gather.GatherAllDemosFromPath(config.BasePath)
    
    if err != nil {
        u.logEvent(fmt.Sprintf("Warning during demo gathering: %v", err))
    }
    
    if len(matches) == 0 {
        u.logEvent("Error: No demo files found or all demos failed to parse")
        return
    }
    
    u.logEvent(fmt.Sprintf("Found %d demos, starting analysis...", len(matches)))
    
    // Process matches
    result, err := wrangle.ProcessMatches(matches, steamIDs)
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
```

## 2. Gather Component (gather.go)

### 2.1 Current State
- Has `GatherDemo()` - analyzes single demo
- Has `GatherAllDemos()` - uses hardcoded `*.dem` pattern in current directory

### 2.2 Required Changes

**Add new function to accept base path:**

```go
func GatherAllDemosFromPath(basePath string) ([]*api.Match, error) {
    // 1. Validate base path exists
    // 2. Walk directory tree recursively
    // 3. Find all .dem files
    // 4. Analyze each demo file
    // 5. Return all matches and aggregated errors
}
```

### 2.3 Implementation Details

**Directory Walking:**
- Use `filepath.WalkDir()` to recursively search subdirectories
- Filter for files with `.dem` extension only (no other formats supported)
- Collect all matching file paths

**Demo Processing:**
- Reuse existing `GatherDemo()` function
- Process each demo file sequentially (or consider parallel processing)
- Aggregate results in `[]*api.Match` slice
- Collect errors but continue processing remaining demos

**Error Handling:**
- Return partial results even if some demos fail
- Use `errors.Join()` to combine multiple errors
- Log each failed demo with filename for debugging

### 2.4 Data Flow

```
basePath (string)
  ↓
Find all .dem files recursively
  ↓
For each demo file:
  - Call api.AnalyzeDemo()
  - Extract Match data
  ↓
Return []*api.Match + errors
```

### 2.5 Example Implementation Pattern

```go
func GatherAllDemosFromPath(basePath string) ([]*api.Match, error) {
    var matches []*api.Match
    var errs []error
    
    // Validate path
    if _, err := os.Stat(basePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("base path does not exist: %s", basePath)
    }
    
    // Walk directory tree
    err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        
        // Skip directories
        if d.IsDir() {
            return nil
        }
        
        // Process .dem files
        if filepath.Ext(path) == ".dem" {
            match, err := GatherDemo(path)
            if err != nil {
                errs = append(errs, fmt.Errorf("%s: %w", path, err))
                return nil // Continue processing other files
            }
            matches = append(matches, match)
        }
        
        return nil
    })
    
    if err != nil {
        errs = append(errs, err)
    }
    
    if len(matches) == 0 {
        return nil, ErrNoDemos
    }
    
    return matches, errors.Join(errs...)
}
```

## 3. Wrangle Component (wrangle.go)

### 3.1 Purpose
Transform raw match data from cs-demo-analyzer into player-specific statistics that can be:
- Filtered by SteamID64
- Grouped by map
- Displayed in tables

### 3.2 Data Structures

Based on STATISTICS.md, create the following structures with support for filtering by map and side (T/CT):

```go
// PlayerStats holds statistics for a single player
type PlayerStats struct {
    SteamID64    string
    PlayerName   string  // For display only
    
    // Per-map statistics with side breakdown
    MapStats map[string]*MapStatistics
    
    // Overall statistics (aggregated across all maps and sides)
    OverallStats *OverallStatistics
}

// MapStatistics holds statistics for a player on a specific map
// Separated by side (T/CT) for filtering
type MapStatistics struct {
    MapName      string
    MatchesPlayed int
    
    // Statistics by side (T = Terrorist, CT = Counter-Terrorist)
    SideStats map[string]*SideStatistics
}

// SideStatistics holds stats for a specific side on a map
type SideStatistics struct {
    Side         string   // "T" or "CT"
    
    // Statistics from STATISTICS.md
    KAST         float64  // Percentage (0-100)
    ADR          float64  // Average Damage per Round
    KD           float64  // Kill/Death ratio
    Kills        int
    Deaths       int
    FirstKills   int
    FirstDeaths  int
    TradeKills   int      // 2-4 second window after teammate death
    TradeDeaths  int      // 2-4 second window after getting a kill
    
    // Additional useful stats
    Assists      int
    Headshots    int
    RoundsPlayed int
}

// OverallStatistics holds aggregated stats across all maps and sides
type OverallStatistics struct {
    // Aggregated statistics from STATISTICS.md
    KAST         float64  // Weighted average
    ADR          float64  // Weighted average
    KD           float64  // Total kills / total deaths
    Kills        int      // Sum
    Deaths       int      // Sum
    FirstKills   int      // Sum
    FirstDeaths  int      // Sum
    TradeKills   int      // Sum
    TradeDeaths  int      // Sum
    
    // Additional stats
    Assists      int
    Headshots    int
    RoundsPlayed int
    MatchesPlayed int
}

// WrangleResult holds the complete analysis result
type WrangleResult struct {
    PlayerStats  []*PlayerStats
    MapList      []string  // Unique maps encountered
    TotalMatches int
}
```

### 3.3 Core Functions

**Main Processing Function:**

```go
func ProcessMatches(matches []*api.Match, steamIDs []string) (*WrangleResult, error) {
    // 1. Create map of steamID -> PlayerStats
    // 2. For each match:
    //    - Extract map name
    //    - For each player in the match:
    //      - Check if player's SteamID is in filter list
    //      - If yes, extract their stats and add to PlayerStats
    // 3. Calculate aggregated stats (KAST, ADR, K/D, etc.)
    // 4. Return WrangleResult
}
```

**Helper Functions:**

```go
// ExtractPlayerFromMatch finds a specific player in a match by SteamID64
// SIMPLIFIED: Use match.PlayersBySteamID[steamID64] directly!
func ExtractPlayerFromMatch(match *api.Match, steamID64 uint64) (*api.Player, bool) {
    player, exists := match.PlayersBySteamID[steamID64]
    return player, exists
}

// NO NEED to implement these - use built-in methods:
// - player.KAST() - KAST is pre-calculated by cs-demo-analyzer
// - player.AverageDamagePerRound() - ADR is pre-calculated
// - player.TradeKillCount() - Trade kills pre-detected (5 second window)
// - player.TradeDeathCount() - Trade deaths pre-detected (5 second window)
// - player.FirstKillCount() - First kills pre-calculated
// - player.FirstDeathCount() - First deaths pre-calculated
// - player.KillCount(), player.DeathCount(), player.KillDeathRatio() - all available

// AggregateStats combines stats from multiple matches into OverallStats
func AggregateStats(mapStats map[string]*MapStatistics) *OverallStatistics

// ExtractStatsBySide analyzes rounds to split stats by T/CT side
// This is the main custom logic needed since player methods return match totals
func ExtractStatsBySide(match *api.Match, player *api.Player, mapStats *MapStatistics)
```

### 3.4 Data Processing Pipeline

```
Input: []*api.Match + []string (SteamIDs to filter)
  ↓
1. Create PlayerStats for each SteamID
  ↓
2. For each Match:
   - Get map name
   - For each round, track which side each player is on
   - For each specified SteamID:
     * Find player in match (if present)
     * Extract raw stats from api.Player for each side
     * Calculate derived stats (KAST, ADR, K/D, trade kills/deaths)
     * Add to PlayerStats.MapStats[mapName].SideStats[side]
  ↓
3. Calculate OverallStats for each player
   - Aggregate across all maps and sides
   - Compute weighted averages for KAST, ADR (weighted by rounds played)
   - Sum totals for kills, deaths, etc.
  ↓
Output: WrangleResult
```

### 3.5 Statistics Calculations

**All statistics from STATISTICS.md are available via cs-demo-analyzer built-in methods:**

**KAST (Kill, Assist, Survive, Trade):**
```go
kast := player.KAST() // Returns float32 percentage (0-100)
```
- Pre-calculated by cs-demo-analyzer
- Logic: (Rounds with K + Rounds with A + Rounds survived + Rounds traded) / Total Rounds
- No need to implement ourselves!

**ADR (Average Damage per Round):**
```go
adr := player.AverageDamagePerRound() // Returns float32
```
- Pre-calculated: Total Health Damage / Rounds Played
- No need to implement ourselves!

**K/D Ratio:**
```go
kd := player.KillDeathRatio() // Returns float32
```
- Pre-calculated: Kills / Deaths (handles division by zero)
- No need to implement ourselves!

**Kills, Deaths, Assists:**
```go
kills := player.KillCount()      // Returns int
deaths := player.DeathCount()    // Returns int
assists := player.AssistCount()  // Returns int
```

**First Kills/Deaths:**
```go
firstKills := player.FirstKillCount()    // Returns int
firstDeaths := player.FirstDeathCount()  // Returns int
```
- Pre-calculated: First kill/death of each round
- No need to implement ourselves!

**Trade Kills/Deaths:**
```go
tradeKills := player.TradeKillCount()    // Returns int
tradeDeaths := player.TradeDeathCount()  // Returns int
```
- Pre-detected by cs-demo-analyzer using 5-second window
- Each Kill object has `IsTradeKill` and `IsTradeDeath` boolean fields
- **Note:** cs-demo-analyzer uses 5-second window (hardcoded), not 2-4 seconds as requested
- **Recommendation:** Use library's 5-second detection and document it

**Bonus Statistics Available:**
```go
headshotCount := player.HeadshotCount()          // Returns int
headshotPercent := player.HeadshotPercent()      // Returns int (0-100)
hltvRating := player.HltvRating()                // Returns float32 (HLTV 1.0)
hltvRating2 := player.HltvRating2()              // Returns float32 (HLTV 2.0)
bombsPlanted := player.BombPlantedCount()        // Returns int
bombsDefused := player.BombDefusedCount()        // Returns int
utilityDamage := player.UtilityDamage()          // Returns int
mvpCount := player.MvpCount                      // int field
```

**Trade Kills/Deaths:**
- Trade Kill: Killing an enemy within 2-4 seconds of them killing a teammate
- Trade Death: Being killed within 2-4 seconds after getting a kill
- Use 2-4 second window as specified

**Implementation Research Required:**
Before implementing KAST and trade kill calculations, research the following:

1. **cs-demo-analyzer Package:**
   - Check if KAST is provided directly in the API
   - Check if trade kills/deaths are calculated by the library
   - Review available player and round statistics
   - Documentation: https://github.com/akiver/cs-demo-analyzer

2. **cs-demo-manager Reference:**
   - Study how cs-demo-manager implements KAST calculation
   - Study how they detect trade kills (timing window, logic)
   - Repository: https://github.com/akiver/cs-demo-manager
   - Look for examples of round event processing

3. **Other Resources:**
   - Search for existing Go implementations of CS:GO statistics
   - Check demoinfocs-golang documentation for event types
   - Look for community implementations and best practices

### 3.6 Implementation Considerations

**cs-demo-analyzer API Usage:**
The `api.Match` struct contains:
- `Players []api.Player` - Player statistics
- `Rounds []api.Round` - Round-by-round data
- `MapName string` - Map identifier
- Team information for determining player sides (T/CT)

The `api.Player` struct likely contains:
- Basic info: Name, SteamID64
- Stats: Kills, Deaths, Assists, Damage, etc.
- Team/side information
- Round-specific data may be in `api.Round` objects

**Important Research Tasks:**
1. Review the cs-demo-analyzer documentation to understand:
   - What statistics are directly available (especially KAST)
   - What needs to be calculated from raw data
   - How to extract round-by-round events for KAST and trade calculations
   - How to determine which side (T/CT) a player was on for each round

2. Examine cs-demo-manager source code:
   - Find their KAST implementation
   - Find their trade kill detection logic
   - Understand their time window calculations
   - See how they handle edge cases

3. Verify round event access:
   - Can we access kill events with timestamps?
   - Can we correlate deaths with team information?
   - How to track if a player survived a round?
   - How to detect if a death was traded?

## 4. Display Component (GUI - Bottom Panel)

### 4.1 Table Structure

The bottom panel should display statistics in tabular format with filtering capabilities:

**Table Layout:**
- Columns: Player | Map | Side | KAST% | ADR | K/D | Kills | Deaths | FK | FD | TK | TD
- Rows grouped by player (sorted alphabetically by name initially)
- Each player has rows for each map they played
- Each map has separate rows for T side and CT side
- Include overall row for each player (when no filters applied)

**Filtering Options:**
The user should be able to filter the table by:
1. **Map filter:** Show only stats from a specific map (e.g., "de_dust2")
2. **Side filter:** Show only T side or only CT side stats
3. **Combined filter:** Both map and side together (e.g., "de_mirage" T side only)
4. **No filter (default):** Show all data with overall stats

**Recommended Implementation:** Single combined table with filter controls

### 4.2 Table Implementation

Using `tview.Table`:

```go
func CreateStatisticsTable(result *WrangleResult) *tview.Table {
    table := tview.NewTable().SetBorders(true).SetFixed(1, 0)
    
    // Header row
    headers := []string{"Player", "Map", "Side", "KAST%", "ADR", "K/D", 
                        "Kills", "Deaths", "FK", "FD", "TK", "TD"}
    
    // Add data rows
    // Sort by player name initially
    // For each player:
    //   - Add rows for each map x side combination
    //   - Add overall row (when no filters)
    
    return table
}
```

### 4.3 Display Features

**Basic Features:**
- Fixed header row
- Scrollable content
- Initial sort by player name (alphabetical)
- Color coding (e.g., green for high KAST, red for low)
- Filter controls for map and side (independent or combined)

**Filter Implementation:**
- Dropdown or input field to select map (or "All")
- Dropdown or buttons to select side: "All", "T", or "CT"
- Apply filters independently or together
- Update table display when filters change

**Advanced Features (Future):**
- Export to CSV
- Filter by map
- Compare players side-by-side
- Highlight best/worst stats

## 5. Integration Flow

### 5.1 Main Execution Flow

```go
// In gui.go

func (u *UI) OnAnalyzeClicked(config AnalysisConfig) {
    // 1. Update status
    u.SetMiddleText("Starting analysis...")
    
    // 2. Validate inputs
    steamIDs := extractSteamIDs(config)
    if len(steamIDs) == 0 {
        u.SetMiddleText("Error: No valid player SteamIDs provided")
        return
    }
    
    // 3. Gather demos
    u.SetMiddleText("Gathering demos from " + config.BasePath + "...")
    matches, err := gather.GatherAllDemosFromPath(config.BasePath)
    if err != nil {
        u.SetMiddleText("Warning: " + err.Error())
    }
    
    if len(matches) == 0 {
        u.SetMiddleText("Error: No demos found")
        return
    }
    
    u.SetMiddleText(fmt.Sprintf("Analyzing %d demos...", len(matches)))
    
    // 4. Process matches
    result, err := wrangle.ProcessMatches(matches, steamIDs)
    if err != nil {
        u.SetMiddleText("Error: " + err.Error())
        return
    }
    
    // 5. Display results
    u.SetMiddleText(fmt.Sprintf("Analysis complete! Found stats for %d players across %d maps", 
                                 len(result.PlayerStats), len(result.MapList)))
    
    // 6. Update bottom panel with table
    u.UpdateBottomPanel(result)
}
```

### 5.2 Error Handling Strategy

**Gather Phase:**
- Continue processing even if some demos fail
- Collect and display errors in status panel
- Show count: "Processed 45/50 demos (5 failed)"

**Wrangle Phase:**
- Handle missing player data gracefully
- Skip players not found in any match
- Display warning if player has no data

**Display Phase:**
- Handle empty results
- Show "No data" message if no matches found

## 6. Implementation Phases

### Phase 1: Basic Functionality (MVP)
1. Implement form in gui.go (5 player inputs + base path)
2. Implement GatherAllDemosFromPath() in gather.go
3. Implement basic PlayerStats and MapStatistics structures in wrangle.go
4. Implement ProcessMatches() with basic stats (kills, deaths, K/D)
5. Display simple table with basic stats

### Phase 2: Complete Statistics
1. Implement KAST calculation
2. Implement ADR calculation
3. Implement First Kills/Deaths tracking
4. Implement Trade Kills/Deaths tracking
5. Update table to show all statistics

### Phase 3: Enhanced UI/UX
1. Add input validation
2. Add progress indicators during analysis
3. Add sorting/filtering to table
4. Add color coding for statistics
5. Add export functionality

### Phase 4: Optimization & Polish
1. Optimize demo processing (parallel processing)
2. Add caching for processed demos
3. Add configuration persistence
4. Improve error messages
5. Add help/documentation in UI

## 7. Technical Considerations

### 7.1 Dependencies

**Already Available:**
- `github.com/rivo/tview` - TUI framework
- `github.com/akiver/cs-demo-analyzer` - Demo analysis
- `github.com/markus-wa/demoinfocs-golang` - Demo parsing (via cs-demo-analyzer)

**May Need:**
- Standard library is sufficient for file operations and data processing

### 7.2 Data Flow Types

```go
// gather.go
type Match = api.Match  // From cs-demo-analyzer

// wrangle.go
type ProcessMatchesInput struct {
    Matches  []*api.Match
    SteamIDs []string
}

type ProcessMatchesOutput struct {
    PlayerStats []*PlayerStats
    MapList     []string
    TotalMatches int
}

// gui.go
type AnalysisConfig struct {
    Players  [5]PlayerInput
    BasePath string
}
```

### 7.3 Threading Considerations

**Current Approach:** Sequential processing
- Simpler to implement and debug
- Acceptable for MVP

**Future Optimization:** Parallel processing
- Use goroutines for demo analysis
- Use worker pool pattern to limit concurrency
- Collect results via channels

### 7.4 Memory Considerations

**Large Demo Collections:**
- Demo files can be large (50-100MB each)
- cs-demo-analyzer loads demos into memory
- Consider processing in batches if memory becomes an issue

**Mitigation Strategies:**
- Process demos one at a time (sequential)
- Don't keep full Match objects in memory after processing
- Extract only needed statistics immediately

### 7.5 Testing Strategy

**Unit Tests:**
- Test KAST calculation with known data
- Test ADR calculation
- Test K/D ratio calculation
- Test data filtering by SteamID

**Integration Tests:**
- Test with sample demo files
- Test with empty directories
- Test with corrupted demos
- Test with missing player data

**Manual Testing:**
- Test UI with various inputs
- Test with real demo collections
- Test error scenarios
- Test with edge cases (0 kills, 0 deaths, etc.)

## 8. File Structure Summary

```
/home/runner/work/manalyzer/manalyzer/
├── src/
│   ├── gui.go          # UI implementation
│   │   - PlayerInput struct
│   │   - AnalysisConfig struct
│   │   - Form creation
│   │   - Button handlers
│   │   - Table display
│   │
│   ├── gather.go       # Demo gathering
│   │   - GatherDemo() [exists]
│   │   - GatherAllDemos() [exists]
│   │   - GatherAllDemosFromPath() [NEW]
│   │
│   └── wrangle.go      # Data processing
│       - PlayerStats struct [NEW]
│       - MapStatistics struct [NEW]
│       - WrangleResult struct [NEW]
│       - ProcessMatches() [NEW]
│       - CalculateKAST() [NEW]
│       - CalculateADR() [NEW]
│       - Helper functions [NEW]
│
├── docs/
│   ├── STATISTICS.md   # Defines required stats
│   ├── STRUCTURE.md    # Overall architecture
│   ├── TECSTACK.md     # Technology choices
│   └── PLAN.md         # This document
│
└── main.go             # Entry point
```

## 9. Example Usage Scenario

**User Workflow:**

1. **Launch Application**
   ```bash
   go run main.go
   ```

2. **Fill in Player Information**
   - Player 1 Name: "ScreaM"
   - Player 1 SteamID64: "76561198033662301"
   - Player 2 Name: "s1mple"
   - Player 2 SteamID64: "76561198034202275"
   - (Leave other players empty)

3. **Set Demo Path**
   - Base Path: "/home/user/csgo/demos"

4. **Click "Analyze"**
   - Status shows: "Gathering demos from /home/user/csgo/demos..."
   - Status shows: "Analyzing 15 demos..."
   - Status shows: "Analysis complete! Found stats for 2 players across 5 maps"

5. **View Results**
   - Bottom panel displays table:
   ```
   ┌────────┬──────────┬───────┬──────┬──────┬───────┬────────┬────┬────┬────┬────┐
   │ Player │   Map    │ KAST% │ ADR  │ K/D  │ Kills │ Deaths │ FK │ FD │ TK │ TD │
   ├────────┼──────────┼───────┼──────┼──────┼───────┼────────┼────┼────┼────┼────┤
   │ScreaM  │ de_dust2 │ 78.5  │ 85.3 │ 1.45 │   87  │   60   │ 12 │  8 │  9 │  5 │
   │ScreaM  │de_mirage │ 72.3  │ 79.1 │ 1.23 │   65  │   53   │  9 │  7 │  7 │  6 │
   │ScreaM  │ Overall  │ 75.4  │ 82.2 │ 1.34 │  152  │  113   │ 21 │ 15 │ 16 │ 11 │
   ├────────┼──────────┼───────┼──────┼──────┼───────┼────────┼────┼────┼────┼────┤
   │s1mple  │ de_dust2 │ 82.1  │ 92.7 │ 1.67 │  102  │   61   │ 15 │  6 │ 11 │  4 │
   │s1mple  │de_mirage │ 79.8  │ 88.4 │ 1.52 │   78  │   51   │ 11 │  8 │  8 │  5 │
   │s1mple  │ Overall  │ 81.0  │ 90.6 │ 1.60 │  180  │  112   │ 26 │ 14 │ 19 │  9 │
   └────────┴──────────┴───────┴──────┴──────┴───────┴────────┴────┴────┴────┴────┘
   ```

## 10. Extension Points for Future Development

### 10.1 Visualization Enhancements
- Add graphs/charts for statistics trends
- Add heatmaps for player positions
- Add timeline view for match progression

### 10.2 Data Export
- Export to CSV
- Export to JSON
- Generate HTML reports
- Export to Excel

### 10.3 Advanced Filtering
- Date range filtering
- Map filtering
- Opponent filtering
- Win/loss filtering

### 10.4 Comparison Features
- Compare two players side-by-side
- Compare player performance across different time periods
- Team statistics (when multiple specified players in same match)

### 10.5 Performance Optimization
- Demo caching (store processed data)
- Incremental updates (only process new demos)
- Database integration for large datasets
- Parallel processing for faster analysis

## 11. Questions Answered

The following clarifications have been provided:

1. **SteamID Format:** ✅ Only SteamID64 format needs to be supported (17 digits)

2. **Player Name Usage:** ✅ Player name is purely for display purposes

3. **Empty Player Slots:** ✅ Users can analyze 1-5 players (not all slots need to be filled)

4. **Demo File Extensions:** ✅ Only support `.dem` files

5. **Error Display:** ✅ Show event log in UI (compact, 3-5 rows) with key events, errors, and warnings

6. **Trade Kill/Death Definition:** ✅ Use 2-4 second window (Note: cs-demo-analyzer uses 5 seconds internally)

7. **KAST Calculation:** ✅ cs-demo-analyzer provides KAST directly via `player.KAST()` method

8. **Table Sorting:** ✅ Initial sort order should be by player name (alphabetical)

9. **Map Grouping:** ✅ Map order doesn't matter for now

10. **Filtering:** ✅ Stats should be filterable by map and side (T/CT) independently or combined

## 11a. cs-demo-analyzer Package Capabilities

After thorough investigation of the cs-demo-analyzer package source code, here are the available features:

### 11a.1 Core API Structure

**Match Object (`api.Match`):**
- Contains complete match data excluding warmup/halftime/after match
- `PlayersBySteamID map[uint64]*Player` - Direct access to players by SteamID64
- `Rounds []*Round` - All round data
- `Kills []*Kill` - All kill events
- `MapName string` - Map identifier
- `TeamA`, `TeamB *Team` - Team information with current sides

**Player Object (`api.Player`):**
The Player struct provides ALL the statistics we need through built-in methods:

```go
// Available Statistics (from STATISTICS.md)
player.KAST() float32                    // ✅ KAST percentage (0-100)
player.AverageDamagePerRound() float32   // ✅ ADR
player.KillDeathRatio() float32          // ✅ K/D ratio
player.KillCount() int                   // ✅ Total kills
player.DeathCount() int                  // ✅ Total deaths
player.FirstKillCount() int              // ✅ First kills per round
player.FirstDeathCount() int             // ✅ First deaths per round
player.TradeKillCount() int              // ✅ Trade kills
player.TradeDeathCount() int             // ✅ Trade deaths

// Bonus Statistics Available
player.AssistCount() int                 // Assists
player.HeadshotCount() int               // Headshot kills
player.HeadshotPercent() int             // Headshot percentage
player.HealthDamage() int                // Total health damage dealt
player.ArmorDamage() int                 // Total armor damage dealt
player.UtilityDamage() int               // Damage from grenades
player.BombPlantedCount() int            // Bombs planted
player.BombDefusedCount() int            // Bombs defused
player.MvpCount int                      // MVP stars
player.HltvRating() float32              // HLTV 1.0 rating
player.HltvRating2() float32             // HLTV 2.0 rating

// Clutch Statistics
player.OneVsOneCount() int               // 1v1 situations
player.OneVsOneWonCount() int            // 1v1 wins
player.OneVsTwoCount() int               // 1v2 situations
// ... up to OneVsFive...

// Multi-kill Statistics
player.OneKillCount() int                // Rounds with 1 kill
player.TwoKillCount() int                // Rounds with 2 kills
player.ThreeKillCount() int              // Rounds with 3 kills
player.FourKillCount() int               // Rounds with 4 kills
player.FiveKillCount() int               // Rounds with 5 kills (ace)

// Player Info
player.SteamID64 uint64                  // Steam ID
player.Name string                       // Player name
player.Team *Team                        // Team reference
```

**Round Object (`api.Round`):**
```go
round.Number int                         // Round number
round.TeamASide common.Team              // Team A's side (T/CT)
round.TeamBSide common.Team              // Team B's side (T/CT)
round.WinnerSide common.Team             // Winning side
round.EndReason events.RoundEndReason    // How round ended
round.Duration int64                     // Round duration in ms
```

**Kill Object (`api.Kill`):**
```go
kill.RoundNumber int                     // Which round
kill.KillerSteamID64 uint64              // Killer's Steam ID
kill.VictimSteamID64 uint64              // Victim's Steam ID
kill.AssisterSteamID64 uint64            // Assister's Steam ID
kill.KillerSide common.Team              // Killer's side (T/CT)
kill.VictimSide common.Team              // Victim's side (T/CT)
kill.IsTradeKill bool                    // ✅ Automatically detected (5 sec window)
kill.IsTradeDeath bool                   // ✅ Automatically detected
kill.IsHeadshot bool                     // Headshot kill
kill.IsThroughSmoke bool                 // Through smoke
kill.WeaponName constants.WeaponName     // Weapon used
kill.Distance float32                    // Kill distance
```

### 11a.2 Key Findings

**1. KAST is Pre-calculated:**
- `player.KAST()` method exists and returns percentage (0-100)
- Implementation logic: tracks kills, assists, survived rounds, and traded deaths per round
- No need to implement KAST calculation ourselves!

**2. Trade Kills/Deaths are Pre-detected:**
- `kill.IsTradeKill` and `kill.IsTradeDeath` boolean fields
- cs-demo-analyzer uses 5-second window (defined in `tradeKillDelaySeconds = 5`)
- User requested 2-4 seconds, but library provides 5 seconds
- **Decision needed:** Use library's 5-second trade detection as-is, or re-implement with 2-4 seconds

**3. First Kills/Deaths are Pre-calculated:**
- `player.FirstKillCount()` and `player.FirstDeathCount()` methods available
- No need to implement first kill detection

**4. ADR is Pre-calculated:**
- `player.AverageDamagePerRound()` returns ADR as float32
- Calculated from `player.HealthDamage() / roundCount`

**5. Side Information Available:**
- Each `Round` has `TeamASide` and `TeamBSide` (common.Team type)
- Each `Kill` has `KillerSide` and `VictimSide`
- `common.Team` values: `TeamTerrorists` (T) or `TeamCounterTerrorists` (CT)
- `Player.Team` references the team, which has `CurrentSide *common.Team`

### 11a.3 Implementation Strategy

**Simplified wrangle.go Approach:**

Since cs-demo-analyzer provides all needed statistics, wrangle.go becomes much simpler:

```go
func ProcessMatches(matches []*api.Match, steamIDs []string) (*WrangleResult, error) {
    playerStatsMap := make(map[uint64]*PlayerStats)
    
    // Initialize PlayerStats for each SteamID
    for _, steamIDStr := range steamIDs {
        steamID64, _ := strconv.ParseUint(steamIDStr, 10, 64)
        playerStatsMap[steamID64] = &PlayerStats{
            SteamID64: steamIDStr,
            MapStats:  make(map[string]*MapStatistics),
        }
    }
    
    // Process each match
    for _, match := range matches {
        mapName := match.MapName
        
        for steamID64, playerStats := range playerStatsMap {
            // Check if player exists in this match
            player, exists := match.PlayersBySteamID[steamID64]
            if !exists {
                continue // Player wasn't in this match
            }
            
            // Set player name (for display)
            if playerStats.PlayerName == "" {
                playerStats.PlayerName = player.Name
            }
            
            // Initialize map stats if needed
            if playerStats.MapStats[mapName] == nil {
                playerStats.MapStats[mapName] = &MapStatistics{
                    MapName:   mapName,
                    SideStats: make(map[string]*SideStatistics),
                }
            }
            
            // Extract stats by side (T/CT)
            // This requires analyzing rounds to determine which side player was on
            extractStatsBySide(match, player, playerStats.MapStats[mapName])
        }
    }
    
    // Calculate overall stats
    for _, playerStats := range playerStatsMap {
        playerStats.OverallStats = calculateOverallStats(playerStats.MapStats)
    }
    
    return &WrangleResult{
        PlayerStats:  playerStatsToSlice(playerStatsMap),
        MapList:      getUniqueMapNames(matches),
        TotalMatches: len(matches),
    }, nil
}

func extractStatsBySide(match *api.Match, player *api.Player, mapStats *MapStatistics) {
    // Need to iterate through rounds and determine which side player was on
    // Then aggregate stats for T side and CT side separately
    
    for _, round := range match.Rounds {
        // Determine player's side for this round
        playerSide := determinePlayerSide(round, player)
        
        sideKey := "T"
        if playerSide == common.TeamCounterTerrorists {
            sideKey = "CT"
        }
        
        // Initialize side stats if needed
        if mapStats.SideStats[sideKey] == nil {
            mapStats.SideStats[sideKey] = &SideStatistics{Side: sideKey}
        }
        
        // For each round, need to extract player's performance
        // This is complex - need to filter kills, deaths, etc. by round
        // May need to store all data and calculate at the end
    }
}
```

**Challenge with Side-specific Stats:**

The Player object's methods return aggregated stats across the entire match, not per-side. To get side-specific stats, we need to:

1. Iterate through each round
2. Determine which side the player was on
3. Filter kills/deaths/damage for that round
4. Aggregate by side

**Alternative Simpler Approach:**

For MVP, we could:
- Show stats per map (not per side initially)
- Use the built-in player methods directly
- Add side filtering in Phase 2

### 11a.4 Recommendations

**Use cs-demo-analyzer's Built-in Statistics:**
- ✅ Use `player.KAST()` directly
- ✅ Use `player.AverageDamagePerRound()` directly
- ✅ Use `player.KillCount()`, `player.DeathCount()` directly
- ✅ Use `player.FirstKillCount()`, `player.FirstDeathCount()` directly
- ✅ Use `player.TradeKillCount()`, `player.TradeDeathCount()` directly
- ✅ Use `player.KillDeathRatio()` directly

**Trade Kill Window:**
- cs-demo-analyzer uses 5 seconds (hardcoded in `tradeKillDelaySeconds`)
- User requested 2-4 seconds
- **Options:**
  1. Use library's 5-second detection as-is (easiest)
  2. Re-implement trade detection with custom window (more work)
  3. Document that library uses 5 seconds
- **Recommendation:** Use library's 5-second window and document it

**Side-specific Statistics:**
- More complex than initial assessment
- Player methods return match-total stats, not per-side
- Need custom aggregation logic to split by T/CT side
- **Recommendation:** Start with per-map stats (Phase 1), add per-side in Phase 2

**Simplified Data Flow:**

```
Match → Player (by SteamID64) → Call built-in methods → Display
```

No need to re-implement KAST, ADR, trade detection, etc.!

## 12. Implementation Priority

**Must Have (Phase 1):**
- [x] 5 player input fields (name + SteamID64)
- [x] Base path input
- [x] Analyze button
- [x] GatherAllDemosFromPath() function
- [x] Basic PlayerStats structure
- [x] Basic statistics: Kills, Deaths, K/D
- [x] Simple table display

**Should Have (Phase 2):**
- [ ] All statistics from STATISTICS.md (KAST, ADR, FK, FD, TK, TD)
- [ ] Map grouping in table
- [ ] Overall statistics row per player
- [ ] Input validation
- [ ] Error handling and display

**Nice to Have (Phase 3+):**
- [ ] Table sorting
- [ ] Color coding
- [ ] Progress indicators
- [ ] Clear button
- [ ] Export functionality

## 13. Success Criteria

The implementation will be considered successful when:

1. ✅ User can input 1-5 player (name + SteamID64 pairs)
2. ✅ User can specify a base directory path
3. ✅ Application recursively finds all .dem files in base path
4. ✅ Application analyzes all found demos
5. ✅ Application filters data to only specified SteamID64s
6. ✅ Application calculates all statistics from STATISTICS.md
7. ✅ Application groups statistics by player and map
8. ✅ Application displays results in table format in bottom panel
9. ✅ Application handles errors gracefully
10. ✅ Application provides status updates during processing

## Conclusion

This plan provides a comprehensive roadmap for implementing the Manalyzer application. The implementation is designed to be:

- **Modular:** Clear separation between GUI, data gathering, and data processing
- **Extensible:** Easy to add new statistics or features
- **Maintainable:** Well-structured data types and clear function responsibilities
- **User-friendly:** Simple form-based input with clear status feedback

The phased approach allows for incremental development, starting with core functionality and gradually adding advanced features. Each phase builds upon the previous one, ensuring a solid foundation before adding complexity.
