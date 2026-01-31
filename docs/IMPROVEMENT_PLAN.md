# Manalyzer Improvement Plan

**Document Purpose:** Research and implementation plan for four major improvements to the Manalyzer CS:GO demo analyzer.

**Status:** Research Phase - For Review

**Created:** 2026-01-31

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Improvement 1: Filtering, Sorting, and Grouping](#improvement-1-filtering-sorting-and-grouping)
4. [Improvement 2: Persistent Storage](#improvement-2-persistent-storage)
5. [Improvement 3: Refactoring](#improvement-3-refactoring)
6. [Improvement 4: Data Visualization](#improvement-4-data-visualization)
7. [Implementation Roadmap](#implementation-roadmap)
8. [Risk Assessment](#risk-assessment)

---

## Executive Summary

This document provides research and recommendations for implementing four major improvements to Manalyzer:

1. **Filtering, Sorting, and Grouping** - Enhance table data manipulation
2. **Persistent Storage** - Save player configurations between sessions
3. **Refactoring** - Improve code organization and maintainability
4. **Data Visualization** - Add graphical charts for player comparison

### Key Recommendations

| Improvement | Approach | Effort | Dependencies | Priority |
|-------------|----------|--------|--------------|----------|
| #1 Filtering/Sorting | Add UI controls for existing filter logic + column sorting | Medium | None | High |
| #2 Persistence | JSON config file in user home directory | Small | None (stdlib) | High |
| #3 Refactoring | Split large files, extract modules | Small | None | High |
| #4 Visualization | **Hybrid: Terminal UI + Web Dashboard with go-echarts** | Medium | go-echarts | Medium |

### Recommended Implementation Order

1. **Phase 1:** Refactoring (set up clean structure)
2. **Phase 2:** Persistent storage (easy win, high value)
3. **Phase 3:** Filtering and sorting (build on clean code)
4. **Phase 4:** Visualization (final feature enhancement)

---

## Current State Analysis

### Codebase Overview

```
manalyzer/
├── main.go           (14 lines)   - Entry point
├── src/
│   ├── gui.go        (605 lines)  - UI components, event handling
│   ├── wrangle.go    (460 lines)  - Statistics extraction and aggregation
│   └── gather.go     (126 lines)  - Demo file discovery and parsing
└── docs/             - Documentation
```

**Total:** ~1,205 lines of Go code

### Technology Stack

- **Language:** Go 1.23+
- **UI Framework:** [tview](https://github.com/rivo/tview) - Terminal UI
- **Terminal Handling:** [tcell](https://github.com/gdamore/tcell)
- **Demo Parser:** [cs-demo-analyzer](https://github.com/akiver/cs-demo-analyzer)

### Current Features

1. ✅ Player statistics tracking (up to 5 players by SteamID64)
2. ✅ Side-specific statistics (T/CT breakdown)
3. ✅ Map-based analysis
4. ✅ Comprehensive metrics (KAST, ADR, K/D, FK/FD, TK/TD)
5. ✅ Recursive demo file scanning
6. ✅ Terminal UI with form input and table display
7. ⚠️ Basic filtering logic exists but NO UI controls
8. ❌ No data persistence
9. ❌ No sorting capability
10. ❌ No visualizations

### Pain Points Identified

1. **User Experience:**
   - Must re-enter player data every session
   - Cannot sort table by different columns
   - Cannot filter data interactively
   - No visual comparison of player performance

2. **Code Organization:**
   - `gui.go` is 605 lines (too large, multiple responsibilities)
   - Some functions in `wrangle.go` exceed 100 lines
   - No separation of concerns between UI rendering and business logic

3. **Functionality Gaps:**
   - Filter logic exists but is unused (no UI)
   - No way to compare players visually
   - No way to spot trends or patterns easily

---

## Improvement 1: Filtering, Sorting, and Grouping

### Current State

**Filtering:**
- ✅ Logic EXISTS in `StatisticsTable.SetFilter()` (gui.go:342-346)
- ✅ Can filter by map name and side (T/CT/Both)
- ❌ NO UI controls to activate these filters
- ❌ No player-specific filtering
- ❌ No stat range filtering (e.g., KAST > 70%)

**Sorting:**
- ✅ Sorts by player name only (hardcoded, gui.go:164-170)
- ❌ Cannot sort by other columns
- ❌ No ascending/descending toggle
- ❌ No sort indicator in UI

**Grouping:**
- ✅ Fixed hierarchy: Player → Map → Side → Stats
- ❌ No alternative grouping options

### Research Findings

#### Filtering Requirements

Users need to filter by:

1. **Map Name** - Show only specific maps (e.g., de_dust2, de_mirage)
2. **Side** - T only, CT only, or Both
3. **Player** - Show/hide specific players
4. **Stat Ranges** - e.g., KAST > 70%, K/D > 1.0, ADR > 75

#### Sorting Requirements

Users need to sort by any column:

| Column | Typical Order | Use Case |
|--------|---------------|----------|
| Player | Alphabetical | Find specific player |
| Map | Alphabetical | Group by map |
| KAST% | Descending | Find best performers |
| ADR | Descending | Find high-damage players |
| K/D | Descending | Find efficient killers |
| Kills | Descending | Find fraggers |
| Deaths | Ascending | Find survivors |
| FK | Descending | Find entry fraggers |
| FD | Ascending | Find who doesn't die first |

#### Grouping Requirements

Alternative grouping strategies:

1. **By Map** (current: by Player)
   - Shows: Map → Player → Side → Stats
   - Use case: Compare all players on same map

2. **By Side**
   - Shows: Side (T/CT) → Player → Map → Stats
   - Use case: Analyze T-side vs CT-side performance

3. **Flat View**
   - Shows: All rows without hierarchy
   - Use case: Maximum data density

### Proposed Solution

#### Phase 1: Add Filter UI (Easy - Logic Exists)

**Add filter controls above the statistics table:**

```
┌─ Filters ────────────────────────────────────────┐
│ Map: [All Maps ▼]  Side: [All ▼]  Player: [All ▼] │
└──────────────────────────────────────────────────┘
┌─ Player Statistics ──────────────────────────────┐
│ [Table data here...]                              │
└──────────────────────────────────────────────────┘
```

**Implementation:**
- Use `tview.DropDown` for Map, Side, Player selection
- Wire dropdowns to existing `SetFilter()` method
- Extract unique map names from data for dropdown options
- Extract player names from data for dropdown options

**Code changes:**
- Modify `newStatisticsTable()` to include filter controls
- Add filter panel (tview.Flex with dropdowns)
- Connect dropdown callbacks to `SetFilter()` and `renderTable()`

**Effort:** Small (2-3 hours)

#### Phase 2: Add Column Sorting (Medium Complexity)

**Make table headers interactive:**

```
┌─ Player Statistics ──────────────────────────────┐
│ Player▼ | Map | Side | KAST%▲ | ADR | K/D | ...  │  <-- Clickable headers
│ Player1 | dust2 | T  | 75.0  | ... | ... | ...  │
└──────────────────────────────────────────────────┘
```

**Implementation approach:**

1. **Add sort state to StatisticsTable:**
```go
type StatisticsTable struct {
    table      *tview.Table
    data       *WrangleResult
    filterMap  string
    filterSide string
    // NEW:
    sortColumn int    // 0=Player, 3=KAST, 4=ADR, 5=K/D, etc.
    sortDesc   bool   // true = descending
}
```

2. **Make headers clickable:**
```go
// In renderTable(), for each header cell:
cell.SetClickedFunc(func() bool {
    st.toggleSort(columnIndex)
    st.renderTable()
    return true
})
```

3. **Implement sorting logic:**
```go
func (st *StatisticsTable) sortData(players []*PlayerStats) {
    sort.Slice(players, func(i, j int) bool {
        switch st.sortColumn {
        case 0: // Player name
            return players[i].PlayerName < players[j].PlayerName
        case 3: // KAST
            return players[i].OverallStats.KAST > players[j].OverallStats.KAST
        case 4: // ADR
            return players[i].OverallStats.ADR > players[j].OverallStats.ADR
        // ... etc for each column
        }
    })
    if st.sortDesc {
        // Reverse order
    }
}
```

4. **Add sort indicator to header:**
```go
header := "KAST%"
if st.sortColumn == 3 {
    header += if st.sortDesc { "▼" } else { "▲" }
}
```

**Code changes:**
- Modify `StatisticsTable` struct
- Modify `renderTable()` to add click handlers
- Add `sortData()` helper function
- Add `toggleSort()` method

**Effort:** Medium (4-6 hours)

#### Phase 3: Advanced Filtering (Optional)

**Add stat range filters:**

```
┌─ Advanced Filters ───────────────────────────────┐
│ KAST: [  ] min  [  ] max   ADR: [  ] min  [  ] max
│ K/D:  [  ] min  [  ] max   [Apply] [Clear]      │
└──────────────────────────────────────────────────┘
```

**Implementation:**
- Add input fields for min/max values
- Create `Filter` struct to hold all criteria
- Apply filters in `renderTable()` before displaying rows

**Effort:** Medium (4-6 hours)

#### Phase 4: Alternative Grouping (Complex - Future)

**Add "Group By" dropdown:**

```
Group by: [Player ▼]  (options: Player, Map, Side, None)
```

**Implementation:**
- Requires significant refactoring of `renderTable()`
- Different rendering logic for each grouping type
- Consider implementing in Phase 3 (Refactoring) to create modular rendering

**Effort:** Large (8-12 hours)

### Recommended Approach

**Implement in stages:**

1. ✅ **Stage 1:** Filter dropdowns (map, side, player) - MUST HAVE
2. ✅ **Stage 2:** Column sorting - SHOULD HAVE
3. ⚠️ **Stage 3:** Stat range filters - NICE TO HAVE
4. ⚠️ **Stage 4:** Alternative grouping - FUTURE

### Testing Strategy

1. Test with no data (empty state)
2. Test with 1 player, 1 map
3. Test with 5 players, 10+ maps (large dataset)
4. Test filter combinations (map + side + player)
5. Test sort on each column
6. Test sort + filter together

---

## Improvement 2: Persistent Storage

### Current State

- ❌ NO persistence mechanism
- ❌ Player data lost when app closes
- ❌ Must re-enter 5 players × (name + SteamID64) each session
- ❌ Must re-enter demo path each session
- ⚠️ Config only exists in memory (`AnalysisConfig` struct)

**Pain point:** For regular users analyzing the same 5 players repeatedly, this is very tedious.

### Research Findings

#### Storage Format Options

| Format | Pros | Cons | Verdict |
|--------|------|------|---------|
| **JSON** | Built-in (`encoding/json`), human-readable, simple | None for this use case | ✅ **RECOMMENDED** |
| TOML | Very readable | Requires `github.com/BurntSushi/toml` | Not needed |
| YAML | Popular for config | Requires third-party lib | Overkill |
| SQLite | Structured queries | Requires CGO or pure-Go lib, overkill | Too complex |
| Binary (gob) | Fast, small | Not human-readable/editable | Bad UX |

**Decision: JSON** - Best balance of simplicity, readability, and Go support.

#### Storage Location Options

| Location | Example | Pros | Cons | Verdict |
|----------|---------|------|------|---------|
| **User config dir** | `~/.config/manalyzer/config.json` | Standard, per-user, persists | Requires `os.UserConfigDir()` | ✅ **RECOMMENDED** |
| User home dir | `~/.manalyzer/config.json` | Simple, visible | Less standard | ⚠️ Fallback |
| Current directory | `./manalyzer-config.json` | Very simple | Lost if running from different dirs | ❌ |
| Executable dir | Same folder as binary | Portable with binary | Non-standard, permission issues | ❌ |

**Decision:** Use `os.UserConfigDir()` with fallback to `os.UserHomeDir()`.

**Paths:**
- Linux: `~/.config/manalyzer/config.json` or `~/.manalyzer/config.json`
- macOS: `~/Library/Application Support/manalyzer/config.json` or `~/.manalyzer/config.json`
- Windows: `%APPDATA%\manalyzer\config.json` or `%USERPROFILE%\.manalyzer\config.json`

### Proposed Solution

#### Config Structure

```go
// config.go

package manalyzer

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// Config represents the persistent application configuration.
type Config struct {
    Version      int            `json:"version"`      // Schema version (for future migrations)
    Players      []PlayerConfig `json:"players"`      // Player list
    LastDemoPath string         `json:"lastDemoPath"` // Last used demo directory
    Preferences  Preferences    `json:"preferences"`  // UI preferences
}

// PlayerConfig stores a player's name and SteamID64.
type PlayerConfig struct {
    Name      string `json:"name"`
    SteamID64 string `json:"steamID64"`
}

// Preferences stores UI and filter preferences.
type Preferences struct {
    DefaultMapFilter  string `json:"defaultMapFilter,omitempty"`  // Default map filter
    DefaultSideFilter string `json:"defaultSideFilter,omitempty"` // Default side filter
    AutoSave          bool   `json:"autoSave"`                    // Auto-save on analyze
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
    return &Config{
        Version:     1,
        Players:     make([]PlayerConfig, 0),
        Preferences: Preferences{AutoSave: true},
    }
}

// configPath returns the path to the config file.
func configPath() (string, error) {
    // Try user config directory first
    configDir, err := os.UserConfigDir()
    if err != nil {
        // Fall back to home directory
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        return filepath.Join(homeDir, ".manalyzer", "config.json"), nil
    }
    return filepath.Join(configDir, "manalyzer", "config.json"), nil
}

// LoadConfig loads configuration from disk.
// Returns default config if file doesn't exist (first run).
func LoadConfig() (*Config, error) {
    path, err := configPath()
    if err != nil {
        return nil, err
    }

    // Check if file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // First run - return default config
        return DefaultConfig(), nil
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

// SaveConfig saves configuration to disk.
func SaveConfig(config *Config) error {
    path, err := configPath()
    if err != nil {
        return err
    }

    // Create directory if it doesn't exist
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Marshal config to JSON (pretty-printed)
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }

    // Write to file
    return os.WriteFile(path, data, 0644)
}
```

#### Example Config File

```json
{
  "version": 1,
  "players": [
    {
      "name": "Player1",
      "steamID64": "76561198012345678"
    },
    {
      "name": "Player2",
      "steamID64": "76561198087654321"
    }
  ],
  "lastDemoPath": "/home/user/csgo/demos",
  "preferences": {
    "defaultMapFilter": "",
    "defaultSideFilter": "",
    "autoSave": true
  }
}
```

#### UI Integration

**Changes to gui.go:**

1. **Load config on startup:**
```go
func New() *UI {
    app := tview.NewApplication()
    
    // Load config
    config, err := LoadConfig()
    if err != nil {
        log.Printf("Failed to load config: %v, using defaults", err)
        config = DefaultConfig()
    }
    
    // Create form with pre-filled values
    form := createPlayerInputForm(config)
    
    // ... rest of initialization
}
```

2. **Pre-fill form fields:**
```go
func createPlayerInputForm(config *Config) *tview.Form {
    form := tview.NewForm()
    
    // Add player inputs (pre-filled from config)
    for i := 0; i < 5; i++ {
        playerName := ""
        steamID := ""
        if i < len(config.Players) {
            playerName = config.Players[i].Name
            steamID = config.Players[i].SteamID64
        }
        
        form.AddInputField(fmt.Sprintf("Player %d Name", i+1), playerName, 30, nil, nil)
        form.AddInputField(fmt.Sprintf("Player %d SteamID64", i+1), steamID, 17, validateSteamID64, nil)
    }
    
    // Add demo path (pre-filled from config)
    form.AddInputField("Demo Base Path", config.LastDemoPath, 50, nil, nil)
    
    // ... buttons
    return form
}
```

3. **Save config when analyzing:**
```go
func (u *UI) onAnalyzeClicked(form *tview.Form) {
    config := u.extractConfigFromForm(form)
    
    // ... validation ...
    
    // Save config if auto-save enabled
    if u.config.Preferences.AutoSave {
        if err := SaveConfig(u.config); err != nil {
            u.logEvent(fmt.Sprintf("Warning: Failed to save config: %v", err))
        } else {
            u.logEvent("Configuration saved")
        }
    }
    
    // Start analysis
    go u.runAnalysis(config)
}
```

4. **Add "Save Config" button (optional):**
```go
form.AddButton("Save Config", func() {
    config := u.extractConfigFromForm(form)
    if err := SaveConfig(config); err != nil {
        u.logEvent(fmt.Sprintf("Error saving config: %v", err))
    } else {
        u.logEvent("Configuration saved successfully")
    }
})
```

### Implementation Steps

1. ✅ Create `src/config.go` with config structures and functions
2. ✅ Add config loading in `New()` function
3. ✅ Modify `createPlayerInputForm()` to accept and use config
4. ✅ Add config saving in `onAnalyzeClicked()`
5. ✅ Add "Save Config" button (optional)
6. ✅ Test first run (no config file)
7. ✅ Test save and reload
8. ✅ Test manual editing of JSON file

### Migration Strategy

**First Release (v1.0 → v1.1):**
- No existing config files, so no migration needed
- Just add config loading/saving

**Future Versions:**
- Use `version` field to detect schema changes
- Implement migration functions:
```go
func migrateConfig(config *Config) *Config {
    switch config.Version {
    case 1:
        // Migrate from v1 to v2
        config = migrateV1ToV2(config)
        fallthrough
    case 2:
        // Migrate from v2 to v3
        config = migrateV2ToV3(config)
    }
    return config
}
```

### Testing Strategy

1. Test first run (no config file)
2. Test saving config
3. Test loading saved config
4. Test with invalid JSON (should use defaults)
5. Test with missing fields (should use defaults)
6. Test with manual edits to JSON
7. Test on Linux, macOS, Windows (different config paths)

### Recommendations

- ✅ Implement as described above
- ✅ Use JSON format
- ✅ Store in user config directory
- ✅ Auto-save on analyze (with opt-out option)
- ⚠️ Consider adding "Reset Config" button to clear saved data
- ⚠️ Consider adding "Export/Import Config" for sharing between machines

**Effort:** Small (3-4 hours)

---

## Improvement 3: Refactoring

### Current State

**File sizes:**
- `gui.go`: 605 lines ⚠️ (too large, multiple responsibilities)
- `wrangle.go`: 460 lines ⚠️ (complex logic, some long functions)
- `gather.go`: 126 lines ✅ (reasonable)
- `main.go`: 14 lines ✅ (simple)

**Code organization issues:**

1. **gui.go** - Multiple responsibilities:
   - EventLog struct and methods
   - StatisticsTable struct and methods
   - Form creation
   - Event handlers
   - Analysis orchestration
   - UI initialization

2. **wrangle.go** - Some functions too long:
   - `extractPlayerStatsBySide()`: 140 lines (does too much)
   - Mixes kill counting, damage aggregation, KAST calculation

3. **No separation** between:
   - UI rendering vs. business logic
   - Data structures vs. data manipulation
   - Config management vs. UI logic

### Research Findings

#### Go Best Practices for Project Structure

**For projects <5000 LOC:**
- Flat package structure is idiomatic
- Split by concern, not by layer
- Files should be <500 lines
- Functions should be <50 lines (ideally <30)

**For projects >5000 LOC:**
- Consider sub-packages
- Group by domain/feature

**Current project: ~1,200 LOC → Will grow to ~2,000 LOC with improvements**
- **Verdict:** Keep flat structure, split files by concern

#### File Naming Conventions

Go convention: `package_aspect.go`

For our case:
- `gui.go` - main UI struct and initialization
- `gui_form.go` - form creation and handlers
- `gui_table.go` - statistics table rendering
- `gui_eventlog.go` - event log component
- `config.go` - configuration management
- `filter.go` - filtering and sorting logic
- `wrangle.go` - main statistics processing
- `wrangle_helpers.go` - helper functions for wrangle
- `gather.go` - demo file gathering (already good)

### Proposed Solution

#### Phase 1: Split gui.go

**Current: gui.go (605 lines)**

**New structure:**

1. **gui.go** (~100 lines) - Main UI orchestration
   ```go
   // UI struct and initialization
   type UI struct { ... }
   func New() *UI
   func (u *UI) Start() error
   func (u *UI) Stop()
   func (u *UI) QueueUpdate(fn func())
   func (u *UI) logEvent(message string)
   ```

2. **gui_form.go** (~150 lines) - Form creation and handlers
   ```go
   func createPlayerInputForm(config *Config) *tview.Form
   func validateSteamID64(text string, lastChar rune) bool
   func (u *UI) setupFormHandlers(form *tview.Form)
   func (u *UI) onAnalyzeClicked(form *tview.Form)
   func (u *UI) onClearClicked(form *tview.Form)
   func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig
   func (u *UI) runAnalysis(config AnalysisConfig)
   ```

3. **gui_table.go** (~250 lines) - Statistics table
   ```go
   type StatisticsTable struct { ... }
   func newStatisticsTable() *StatisticsTable
   func (st *StatisticsTable) UpdateData(result *WrangleResult)
   func (st *StatisticsTable) renderTable()
   func (st *StatisticsTable) addDataRow(...)
   func (st *StatisticsTable) addMapSummaryRow(...)
   func (st *StatisticsTable) addOverallRow(...)
   func (st *StatisticsTable) SetFilter(mapFilter, sideFilter string)
   // NEW: Sorting functions (Improvement #1)
   func (st *StatisticsTable) sortData(...)
   func (st *StatisticsTable) toggleSort(column int)
   ```

4. **gui_eventlog.go** (~80 lines) - Event log component
   ```go
   type EventLog struct { ... }
   func newEventLog(maxLines int) *EventLog
   func (el *EventLog) Log(message string)
   func (el *EventLog) LogError(message string)
   ```

**Benefits:**
- Each file <300 lines
- Clear separation of concerns
- Easier to navigate and maintain
- Easier to test individual components

**Implementation:**
- Simple file split (copy-paste + organize imports)
- All functions stay in same package
- No breaking changes to API

**Effort:** Small (1-2 hours)

#### Phase 2: Extract Config Management

**New file: config.go** (~150 lines)

Already covered in Improvement #2.

Includes:
- Config struct
- LoadConfig()
- SaveConfig()
- DefaultConfig()
- configPath()

**Effort:** Small (covered in Improvement #2)

#### Phase 3: Extract Filter Logic

**New file: filter.go** (~100 lines)

```go
package manalyzer

// Filter represents data filtering criteria.
type Filter struct {
    MapName    string
    Side       string
    PlayerIDs  []string
    MinKAST    float64
    MaxKAST    float64
    MinADR     float64
    MaxADR     float64
    MinKD      float64
    MaxKD      float64
}

// Apply filters the WrangleResult based on criteria.
func (f *Filter) Apply(data *WrangleResult) *WrangleResult {
    filtered := &WrangleResult{
        PlayerStats:  make([]*PlayerStats, 0),
        MapList:      data.MapList,
        TotalMatches: data.TotalMatches,
    }
    
    for _, player := range data.PlayerStats {
        if f.matchesPlayer(player) {
            filtered.PlayerStats = append(filtered.PlayerStats, player)
        }
    }
    
    return filtered
}

func (f *Filter) matchesPlayer(player *PlayerStats) bool {
    // Apply all filter criteria
    // ...
}

// Sorter handles sorting logic.
type Sorter struct {
    Column     string // "player", "kast", "adr", "kd", etc.
    Descending bool
}

// Sort sorts the WrangleResult by the specified column.
func (s *Sorter) Sort(data *WrangleResult) {
    sort.Slice(data.PlayerStats, func(i, j int) bool {
        return s.compare(data.PlayerStats[i], data.PlayerStats[j])
    })
}

func (s *Sorter) compare(a, b *PlayerStats) bool {
    // Comparison logic for each column
    // ...
}
```

**Benefits:**
- Filtering logic separate from UI
- Can be tested independently
- Reusable across different interfaces (future CLI?)
- Cleaner UI code

**Effort:** Medium (2-3 hours)

#### Phase 4: Refactor Large Functions in wrangle.go

**Current issue: `extractPlayerStatsBySide()` is 140 lines**

Break down into smaller focused functions:

```go
// BEFORE: extractPlayerStatsBySide() - 140 lines

// AFTER: Multiple focused functions
func extractPlayerStatsBySide(match *api.Match, player *api.Player) map[string]*SideStatistics {
    sideStats := initializeSideStats()
    
    countRoundsPlayed(match, player, sideStats)
    countKillsAndDeaths(match, player, sideStats)
    countFirstKillsAndDeaths(match, player, sideStats)
    aggregateDamage(match, player, sideStats)
    calculateKDRatios(sideStats)
    calculateKAST(match, player, sideStats)
    
    return sideStats
}

// Each helper function is <30 lines, focused, testable
func initializeSideStats() map[string]*SideStatistics { ... }
func countRoundsPlayed(match, player, sideStats) { ... }
func countKillsAndDeaths(match, player, sideStats) { ... }
func countFirstKillsAndDeaths(match, player, sideStats) { ... }
func aggregateDamage(match, player, sideStats) { ... }
func calculateKDRatios(sideStats) { ... }
func calculateKAST(match, player, sideStats) { ... }
```

**Benefits:**
- Each function <30 lines
- Easier to understand
- Easier to test
- Easier to optimize individually
- Better code reuse

**Effort:** Medium (3-4 hours)

#### Phase 5: Consider Sub-Packages (Future)

**Only if project grows beyond 3,000 LOC**

```
manalyzer/
├── main.go
├── ui/        # All GUI code
│   ├── ui.go
│   ├── form.go
│   ├── table.go
│   └── eventlog.go
├── stats/     # Statistics processing
│   ├── wrangle.go
│   └── helpers.go
├── demo/      # Demo file handling
│   └── gather.go
├── config/    # Configuration
│   └── config.go
└── filter/    # Filtering and sorting
    └── filter.go
```

**Current verdict:** NOT YET - current size doesn't justify this complexity.

### Recommended Approach

**Implement phases in order:**

1. ✅ **Phase 1:** Split gui.go (easy, high value)
2. ✅ **Phase 2:** Extract config.go (covered in Improvement #2)
3. ✅ **Phase 3:** Extract filter.go (needed for Improvement #1)
4. ⚠️ **Phase 4:** Refactor large functions (nice to have)
5. ❌ **Phase 5:** Sub-packages (not needed yet)

### File Structure After Refactoring

```
manalyzer/
├── main.go              (14 lines)
├── src/
│   ├── gui.go           (~100 lines) - UI orchestration
│   ├── gui_form.go      (~150 lines) - Form and handlers
│   ├── gui_table.go     (~250 lines) - Statistics table
│   ├── gui_eventlog.go  (~80 lines)  - Event log
│   ├── config.go        (~150 lines) - Config management
│   ├── filter.go        (~100 lines) - Filtering/sorting
│   ├── wrangle.go       (~400 lines) - Stats processing
│   └── gather.go        (~126 lines) - Demo gathering
└── docs/
```

**Total:** ~1,370 lines (before improvements) → ~2,000 lines (after all improvements)

### Testing Strategy

1. Ensure all tests pass after each refactoring step
2. Verify UI behavior unchanged
3. Check import cycles (Go will catch this)
4. Verify no performance regression

### Recommendations

- ✅ Implement Phases 1-3 (essential)
- ⚠️ Consider Phase 4 if time permits
- ❌ Skip Phase 5 for now

**Effort:** Small to Medium (4-6 hours total)

---

## Improvement 4: Data Visualization

### Current State

- ✅ Text-based table display (tview.Table)
- ✅ Terminal UI (no graphics)
- ❌ No visual comparison of player performance
- ❌ No charts or graphs
- ❌ Hard to spot trends or patterns

**Problem:** tview does NOT support graphical charts/plots. It's terminal-only.

### Research Findings

#### Visualization Use Cases

Users want to:

1. **Compare players side-by-side** - Bar charts showing KAST, ADR, K/D
2. **Analyze T vs CT performance** - Grouped bar charts
3. **See correlations** - Scatter plots (e.g., KAST vs K/D)
4. **Compare across maps** - Heatmaps or grouped bars
5. **Identify strengths/weaknesses** - Radar/spider charts

#### Option 1: Terminal-Based ASCII Charts

**Libraries:**
- [termui](https://github.com/gizak/termui) - TUI widgets (bar, line, gauge, spark)
- [asciigraph](https://github.com/guptarohit/asciigraph) - Simple ASCII line charts

**Example with termui:**
```
┌─ Player Comparison ───────────────────┐
│                                        │
│  KAST%  ████████████ Player1 (75.0)   │
│         ████████ Player2 (60.0)        │
│         ██████████ Player3 (70.0)      │
│                                        │
│  ADR    ██████████████ Player1 (85.2)  │
│         █████████ Player2 (68.5)       │
│         ███████████ Player3 (79.1)     │
└────────────────────────────────────────┘
```

**Pros:**
- Stays in terminal (no browser needed)
- Lightweight
- Consistent with terminal UI

**Cons:**
- Limited chart types (bar, line, gauge only)
- ASCII graphics less clear than true charts
- termui and tview both use tcell - WILL CONFLICT
- Cannot run simultaneously without complex coordination
- Poor readability for detailed data

**Verdict:** ❌ Not recommended - conflicts with tview, limited functionality

#### Option 2: Web-Based Dashboard (RECOMMENDED)

**Libraries:**
- [go-echarts](https://github.com/go-echarts/go-echarts) ✅ **TOP CHOICE**
  - Pure Go, no JavaScript coding needed
  - Built on Apache ECharts (professional charting library)
  - Generates HTML with interactive charts
  - Supports: bar, line, pie, scatter, radar, heatmap, box, funnel, gauge, etc.
  - Can export to PNG/SVG
  - Active development, well-maintained

**Architecture:**

```
┌──────────────────────────────────────────────────────────────┐
│                  Terminal UI (tview)                         │
│  ┌────────────────┐  ┌──────────────────────────────┐       │
│  │ Form Input     │  │ Event Log                    │       │
│  │                │  │                              │       │
│  │ [Analyze]      │  │ Statistics Table             │       │
│  │ [Visualize] ← ─┼─►│                              │       │
│  └────────────────┘  └──────────────────────────────┘       │
└──────────────────────────────────────────────────────────────┘
                               │
                               │ Button click
                               ▼
                    ┌──────────────────────┐
                    │  HTTP Server         │
                    │  :8080               │
                    │  (go-echarts)        │
                    └──────────────────────┘
                               │
                               │ Serves HTML/JS
                               ▼
                    ┌──────────────────────┐
                    │  Web Browser         │
                    │  (auto-opens)        │
                    │                      │
                    │  [Interactive        │
                    │   Charts]            │
                    └──────────────────────┘
```

**Implementation approach:**

1. **Add "Visualize" button to GUI**
2. **Start HTTP server in background** (goroutine)
3. **Generate chart pages** with go-echarts
4. **Auto-open browser** to view charts
5. **Keep terminal UI running** (dual-interface)

**Example chart generation:**

```go
// src/visualize.go

package manalyzer

import (
    "net/http"
    "github.com/go-echarts/go-echarts/v2/charts"
    "github.com/go-echarts/go-echarts/v2/opts"
)

// StartVisualizationServer starts HTTP server for chart visualization.
func StartVisualizationServer(data *WrangleResult, port string) error {
    http.HandleFunc("/", dashboardHandler(data))
    http.HandleFunc("/player-comparison", playerComparisonHandler(data))
    http.HandleFunc("/map-breakdown", mapBreakdownHandler(data))
    http.HandleFunc("/side-performance", sidePerformanceHandler(data))
    
    return http.ListenAndServe(port, nil)
}

func playerComparisonHandler(data *WrangleResult) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        bar := charts.NewBar()
        bar.SetGlobalOptions(
            charts.WithTitleOpts(opts.Title{Title: "Player Comparison"}),
            charts.WithTooltipOpts(opts.Tooltip{Show: true}),
            charts.WithLegendOpts(opts.Legend{Show: true}),
        )
        
        // Extract player names
        players := make([]string, len(data.PlayerStats))
        kastData := make([]opts.BarData, len(data.PlayerStats))
        adrData := make([]opts.BarData, len(data.PlayerStats))
        
        for i, player := range data.PlayerStats {
            players[i] = player.PlayerName
            kastData[i] = opts.BarData{Value: player.OverallStats.KAST}
            adrData[i] = opts.BarData{Value: player.OverallStats.ADR}
        }
        
        bar.SetXAxis(players)
        bar.AddSeries("KAST%", kastData)
        bar.AddSeries("ADR", adrData)
        
        bar.Render(w)
    }
}
```

**Chart types to implement:**

1. **Player Comparison (Bar Chart)**
   - X-axis: Player names
   - Y-axis: Stats (KAST, ADR, K/D)
   - Multi-series (one bar per stat per player)
   - Use case: See who performs best overall

2. **T vs CT Performance (Grouped Bar Chart)**
   - X-axis: Players
   - Y-axis: Stats
   - Two bars per player (T side vs CT side)
   - Use case: Identify if player is T-sided or CT-sided

3. **Map Breakdown (Heatmap)**
   - X-axis: Maps
   - Y-axis: Players
   - Color intensity: Performance metric (KAST or K/D)
   - Use case: Find best maps for each player

4. **Stat Correlation (Scatter Plot)**
   - X-axis: KAST%
   - Y-axis: K/D ratio
   - Each point: One player
   - Use case: See if high KAST correlates with high K/D

5. **Player Radar (Spider/Radar Chart)**
   - Axes: KAST, ADR, K/D, FK, Headshot %
   - Multiple players overlaid
   - Use case: Compare overall player profiles

**Dashboard layout:**

```html
<!DOCTYPE html>
<html>
<head>
    <title>Manalyzer - Player Statistics Dashboard</title>
</head>
<body>
    <h1>CS:GO Player Statistics</h1>
    
    <div class="chart-grid">
        <div id="player-comparison" style="width: 800px; height: 400px;"></div>
        <div id="side-performance" style="width: 800px; height: 400px;"></div>
        <div id="map-breakdown" style="width: 800px; height: 400px;"></div>
        <div id="stat-correlation" style="width: 400px; height: 400px;"></div>
        <div id="player-radar" style="width: 400px; height: 400px;"></div>
    </div>
    
    <!-- Charts rendered by go-echarts -->
</body>
</html>
```

**Browser auto-open:**

```go
import (
    "os/exec"
    "runtime"
)

func openBrowser(url string) error {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "linux":
        cmd = exec.Command("xdg-open", url)
    case "darwin": // macOS
        cmd = exec.Command("open", url)
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", url)
    default:
        return fmt.Errorf("unsupported platform")
    }
    
    return cmd.Start()
}
```

**UI integration:**

```go
// In gui_form.go

form.AddButton("Visualize", func() {
    if u.statsTable.data == nil {
        u.logEvent("Error: No data to visualize. Run analysis first.")
        return
    }
    
    // Start visualization server
    go func() {
        if err := StartVisualizationServer(u.statsTable.data, ":8080"); err != nil {
            u.logEvent(fmt.Sprintf("Error starting visualization server: %v", err))
        }
    }()
    
    // Wait a moment for server to start
    time.Sleep(500 * time.Millisecond)
    
    // Open browser
    if err := openBrowser("http://localhost:8080"); err != nil {
        u.logEvent(fmt.Sprintf("Error opening browser: %v", err))
        u.logEvent("Visit http://localhost:8080 manually to view charts")
    } else {
        u.logEvent("Visualization dashboard opened in browser")
    }
})
```

**Pros:**
- Rich, interactive, professional-quality charts
- No terminal limitations
- go-echarts is mature and actively maintained
- Pure Go (no external JS dependencies to manage)
- Can show multiple charts simultaneously
- Charts are interactive (zoom, hover tooltips, export)
- Keep terminal UI for table data (best of both worlds)
- Can view charts side-by-side with terminal

**Cons:**
- Requires web browser
- Slightly more complex than terminal-only
- Need to handle port conflicts (use dynamic port or config)
- HTTP server runs until app exits

**Mitigations:**
- Check if port is available, try alternatives
- Add config option to disable visualization
- Gracefully handle browser open failure
- Provide manual URL in event log

#### Option 3: Export Static Images

**Libraries:**
- [gonum/plot](https://github.com/gonum/plot) - Generate PNG/SVG charts

**Approach:**
- Generate PNG files on disk
- Open with system image viewer
- Or export CSV for Excel/LibreOffice

**Pros:**
- No runtime dependencies (after generation)
- Professional quality
- Shareable files

**Cons:**
- Not interactive during analysis
- Extra step for user
- gonum/plot has steep learning curve
- Less useful than interactive charts

**Verdict:** ⚠️ Fallback option if web dashboard is rejected

#### Option 4: No Visualization

**Keep terminal UI only, skip charts entirely**

**Pros:**
- Simple, no new dependencies
- Terminal-only is lightweight

**Cons:**
- Misses key requirement
- Users need visual comparison for effective analysis
- Table-only is tedious for spotting patterns

**Verdict:** ❌ Does not meet requirements

### Recommended Approach

**STRONGLY RECOMMEND: Option 2 - Web Dashboard with go-echarts**

#### Implementation Plan

**Phase 1: Basic infrastructure**
1. Add `github.com/go-echarts/go-echarts/v2` dependency
2. Create `src/visualize.go`
3. Implement HTTP server with basic dashboard page
4. Add "Visualize" button to GUI
5. Implement browser auto-open

**Phase 2: Chart implementations**
1. Player comparison bar chart
2. T vs CT performance grouped bar chart
3. Map breakdown heatmap
4. Stat correlation scatter plot
5. Player radar chart

**Phase 3: Polish**
1. Add chart export functionality (PNG/SVG)
2. Add port configuration option
3. Improve dashboard styling (CSS)
4. Add chart filtering (show/hide players)
5. Add graceful server shutdown

### File Structure

```
src/
├── visualize.go          (~300 lines) - HTTP server and chart generation
├── visualize_charts.go   (~200 lines) - Individual chart implementations
```

### Dependencies

```go
// go.mod
require (
    github.com/go-echarts/go-echarts/v2 v2.3.3
)
```

### Testing Strategy

1. Test with no data (should show error)
2. Test with 1 player (edge case)
3. Test with 5 players, multiple maps
4. Test port conflict handling
5. Test on Linux, macOS, Windows
6. Test chart interactivity (zoom, tooltips)
7. Test with browser not available
8. Test concurrent access (multiple browser tabs)

### Alternative Considerations

**If web dashboard is rejected:**
- Implement Option 3 (static PNG export with gonum/plot)
- Or skip visualization entirely

**If go-echarts is rejected:**
- Consider plotly (requires more JS knowledge)
- Or use gonum/plot (steeper learning curve)

### Recommendations

- ✅ **Implement Option 2** (web dashboard with go-echarts)
- ✅ Keep terminal UI as primary interface
- ✅ Make visualization optional (button-activated)
- ✅ Implement 5 chart types (bar, grouped bar, heatmap, scatter, radar)
- ⚠️ Consider adding export functionality
- ⚠️ Consider adding chart customization options

**Effort:** Medium (6-8 hours)

---

## Implementation Roadmap

### Recommended Sequence

#### Phase 1: Foundation and Refactoring (Week 1)

**Goal:** Clean up codebase, establish structure for new features

1. **Refactoring (Improvement #3)** - 4-6 hours
   - Split gui.go into 4 files (gui.go, gui_form.go, gui_table.go, gui_eventlog.go)
   - Verify all functionality still works
   - Update imports

2. **Persistent Storage (Improvement #2)** - 3-4 hours
   - Create config.go
   - Implement LoadConfig() and SaveConfig()
   - Modify GUI to pre-fill from config
   - Add auto-save on analyze
   - Test save/load cycle

**Deliverables:**
- ✅ Cleaner file structure
- ✅ Config persistence working
- ✅ All existing features still work

**Testing:**
- Unit tests for config.go
- Integration test: save → exit → restart → load
- Manual test on Linux, macOS, Windows

#### Phase 2: Enhanced Table Features (Week 2)

**Goal:** Add filtering and sorting to improve data exploration

3. **Filtering UI (Improvement #1, Part 1)** - 2-3 hours
   - Create filter.go
   - Add dropdown filters above table (Map, Side, Player)
   - Wire dropdowns to existing SetFilter() logic
   - Test filter combinations

4. **Column Sorting (Improvement #1, Part 2)** - 4-6 hours
   - Add sort state to StatisticsTable
   - Make table headers clickable
   - Implement sortData() function
   - Add sort indicators (▲/▼) to headers
   - Test sorting on all columns

5. **Advanced Filtering (Optional, Improvement #1, Part 3)** - 4-6 hours
   - Add stat range filters (min/max inputs)
   - Implement range filtering logic
   - Add Apply/Clear buttons
   - Test complex filter scenarios

**Deliverables:**
- ✅ Filter dropdowns working
- ✅ Column sorting working
- ⚠️ (Optional) Stat range filters working

**Testing:**
- Test each filter independently
- Test filter combinations
- Test sorting with filters active
- Test with large dataset (50+ rows)

#### Phase 3: Visualization (Week 3-4)

**Goal:** Add visual comparison of player statistics

6. **Visualization Infrastructure (Improvement #4, Part 1)** - 2-3 hours
   - Add go-echarts dependency
   - Create visualize.go
   - Implement HTTP server
   - Add "Visualize" button
   - Implement browser auto-open
   - Create basic dashboard page

7. **Chart Implementations (Improvement #4, Part 2)** - 4-5 hours
   - Implement player comparison bar chart
   - Implement T vs CT grouped bar chart
   - Implement map breakdown heatmap
   - Implement stat correlation scatter plot
   - Implement player radar chart
   - Create unified dashboard layout

8. **Visualization Polish (Optional)** - 2-3 hours
   - Add chart export (PNG/SVG)
   - Add port configuration
   - Improve dashboard styling
   - Add chart filtering options
   - Implement graceful server shutdown

**Deliverables:**
- ✅ Working visualization server
- ✅ 5 interactive chart types
- ⚠️ (Optional) Chart export and customization

**Testing:**
- Test each chart type
- Test on different browsers (Chrome, Firefox, Safari)
- Test on Linux, macOS, Windows
- Test with no data, 1 player, 5 players
- Test port conflict handling
- Test browser not available scenario

### Milestones

| Milestone | Completion | Features |
|-----------|------------|----------|
| **M1: Refactored** | End of Week 1 | Clean code structure, config persistence |
| **M2: Enhanced Table** | End of Week 2 | Filtering, sorting, better UX |
| **M3: Visualizations** | End of Week 3 | Interactive charts |
| **M4: Polish** | End of Week 4 | Export, customization, docs |

### Estimated Total Effort

| Improvement | Effort | Priority |
|-------------|--------|----------|
| #1 Filtering/Sorting | 6-9 hours | High |
| #2 Persistence | 3-4 hours | High |
| #3 Refactoring | 4-6 hours | High |
| #4 Visualization | 6-8 hours | Medium |
| **Total Core** | **19-27 hours** | **~3-4 weeks** |
| **Total with Optional** | **27-36 hours** | **~4-5 weeks** |

### Dependencies Between Improvements

```
Phase 1: Refactoring (#3) + Persistence (#2)
   │
   ├─► Clean structure for new features
   │
Phase 2: Filtering/Sorting (#1)
   │
   ├─► Uses filter.go from refactoring
   │
Phase 3: Visualization (#4)
   │
   └─► Independent, uses existing data structures
```

**Note:** Phases 2 and 3 can be developed in parallel by different developers.

---

## Risk Assessment

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **tview version conflicts** | Low | Medium | Lock tview version, test upgrades carefully |
| **Browser not available** | Low | Low | Provide manual URL in event log |
| **Port conflicts (8080 in use)** | Medium | Low | Implement dynamic port selection, config option |
| **Config file corruption** | Low | Medium | Validate JSON on load, use defaults on error |
| **Performance with large datasets** | Low | Medium | Test with 100+ matches, optimize if needed |
| **Cross-platform issues** | Low | Medium | Test on Linux, macOS, Windows |

### Implementation Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Scope creep** | Medium | High | Stick to defined phases, defer optional features |
| **Breaking existing features** | Low | High | Comprehensive testing after each phase |
| **Overly complex filtering UI** | Medium | Medium | Start simple (dropdowns only), iterate if needed |
| **Poor chart readability** | Low | Medium | Use go-echarts defaults (already good), iterate based on feedback |

### User Experience Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Confusing filter UI** | Medium | Medium | Follow standard patterns, add tooltips |
| **Too many visualization options** | Low | Low | Start with 5 charts, add more based on feedback |
| **Config file location unclear** | Low | Low | Add config file path to help text / README |
| **Unexpected config overwrites** | Medium | Medium | Add "Reset Config" button, document auto-save behavior |

### Mitigation Strategy Summary

1. **Test incrementally** - Test after each phase, not just at the end
2. **Maintain compatibility** - Don't break existing features
3. **Provide fallbacks** - Handle errors gracefully (missing config, browser not available, etc.)
4. **Document thoroughly** - Update README with new features and config file location
5. **Get feedback early** - Share screenshots/demos after Phase 2

---

## Success Criteria

### Improvement #1: Filtering, Sorting, Grouping

✅ **Minimum Viable:**
- [ ] Dropdown filters for Map, Side, Player working
- [ ] Column sorting on all stat columns
- [ ] Sort direction indicator visible (▲/▼)
- [ ] Filters and sorting work together

⚠️ **Nice to Have:**
- [ ] Stat range filters (min/max)
- [ ] Alternative grouping views

### Improvement #2: Persistent Storage

✅ **Minimum Viable:**
- [ ] Config file saved in user home directory
- [ ] Player names and SteamID64s persist between sessions
- [ ] Demo path persists
- [ ] Auto-load on startup
- [ ] Auto-save on analyze
- [ ] Handles first run (no config file) gracefully

⚠️ **Nice to Have:**
- [ ] "Save Config" button
- [ ] "Reset Config" button
- [ ] Export/Import config

### Improvement #3: Refactoring

✅ **Minimum Viable:**
- [ ] gui.go split into 4 files (<300 lines each)
- [ ] config.go extracted
- [ ] filter.go extracted
- [ ] All existing features work identically
- [ ] No performance regression

⚠️ **Nice to Have:**
- [ ] Large functions in wrangle.go refactored
- [ ] Comprehensive unit tests for extracted modules

### Improvement #4: Data Visualization

✅ **Minimum Viable:**
- [ ] "Visualize" button added to GUI
- [ ] HTTP server starts on button click
- [ ] Browser auto-opens (or manual URL provided)
- [ ] 3+ chart types implemented and working
- [ ] Charts are interactive (hover, zoom)

⚠️ **Nice to Have:**
- [ ] 5+ chart types
- [ ] Chart export (PNG/SVG)
- [ ] Chart customization options
- [ ] Graceful server shutdown

### Overall Success

✅ **User can:**
- [ ] Enter player data once, reuse it across sessions
- [ ] Filter and sort table data to find insights
- [ ] Visualize player comparisons in interactive charts
- [ ] Navigate clean, maintainable codebase

---

## Next Steps

### Before Implementation

1. **Review this plan** with stakeholders
2. **Get approval** on approach for each improvement
3. **Clarify priorities** - Must-have vs. Nice-to-have
4. **Set up development environment**
5. **Create feature branch**

### During Implementation

1. **Follow phased approach** (don't skip ahead)
2. **Test after each change**
3. **Document as you go** (update README)
4. **Commit frequently** with clear messages
5. **Take screenshots** of UI changes

### After Implementation

1. **Comprehensive testing** on all platforms
2. **Update documentation** (README, usage guide)
3. **Create demo video/screenshots**
4. **Gather user feedback**
5. **Plan next iteration** based on feedback

---

## Appendix A: Alternative Approaches Considered

### Filtering/Sorting

**Alternative 1: Command-line flags**
- `--filter-map=de_dust2 --sort-by=kast`
- Rejected: Less interactive, not suitable for TUI

**Alternative 2: Query language**
- Custom mini-language: `map:dust2 AND kast>70`
- Rejected: Overengineered for this use case

### Persistence

**Alternative 1: Environment variables**
- `MANALYZER_PLAYERS="player1:steamid,player2:steamid"`
- Rejected: Not user-friendly, hard to manage 5 players

**Alternative 2: Command-line arguments**
- `./manalyzer --player1-name=Name --player1-steam=ID`
- Rejected: Too verbose, not persistent

### Visualization

**Alternative 1: Terminal-only with termui**
- Rejected: Conflicts with tview, limited chart types

**Alternative 2: Electron app**
- Full GUI with web technologies
- Rejected: Too heavy, loses terminal simplicity

**Alternative 3: Export to external tools**
- Generate CSV, open in Excel
- Rejected: Extra steps, less convenient

---

## Appendix B: Go Library Research

### Charting Libraries Evaluated

| Library | Type | Pros | Cons | Verdict |
|---------|------|------|------|---------|
| **go-echarts** | Web | Interactive, pure Go, feature-rich | Requires browser | ✅ **Selected** |
| gonum/plot | Static image | High quality, no browser | Not interactive, steep curve | ⚠️ Fallback |
| termui | Terminal | Stays in terminal | Conflicts with tview | ❌ |
| asciigraph | Terminal | Simple | Very limited | ❌ |
| plotly-go | Web | Very powerful | More complex than go-echarts | ⚠️ Alternative |
| chart | Static image | Simple API | Limited features | ❌ |

### Config Libraries Evaluated

| Library | Format | Pros | Cons | Verdict |
|---------|--------|------|------|---------|
| **encoding/json** | JSON | Built-in, simple, readable | None | ✅ **Selected** |
| BurntSushi/toml | TOML | Very readable | Extra dependency | ⚠️ Alternative |
| go-yaml/yaml | YAML | Popular | Extra dependency | ❌ |
| spf13/viper | Multi-format | Feature-rich | Overkill | ❌ |

---

## Appendix C: References

### Documentation

- [tview Tutorial](https://github.com/rivo/tview/wiki/Tutorial)
- [go-echarts Examples](https://go-echarts.github.io/go-echarts/)
- [cs-demo-analyzer API](https://github.com/akiver/cs-demo-analyzer)

### Related Projects

- [csgo-stats](https://github.com/markus-wa/demoinfocs-golang) - Demo analysis examples
- [csgostats.gg](https://csgostats.gg/) - Web-based stats (inspiration)

### Standards

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)

---

**END OF DOCUMENT**
