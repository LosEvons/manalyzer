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
    Name     string
    SteamID64 string
}

type AnalysisConfig struct {
    Players  [5]PlayerInput
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
│   Player     │   Status/Log   │
│   Input      │   TextView     │
│   Form       ├────────────────┤
│              │   Statistics   │
│   [Analyze]  │   Table(s)     │
│   [Clear]    │                │
└──────────────┴────────────────┘
```

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

3. **Wire Button Handlers**
   - "Analyze" button → Collect form data → Call analysis pipeline
   - "Clear" button → Reset all form fields

4. **Update Bottom Panel for Tables**
   - Replace Bottom Box with `tview.Table` or multiple tables
   - Support pagination/scrolling for large datasets

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
- Filter for files with `.dem` extension
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

Based on STATISTICS.md, create the following structures:

```go
// PlayerStats holds statistics for a single player
type PlayerStats struct {
    SteamID64    string
    PlayerName   string
    
    // Per-map statistics
    MapStats map[string]*MapStatistics
    
    // Overall statistics (aggregated across all maps)
    OverallStats *MapStatistics
}

// MapStatistics holds statistics for a player on a specific map
type MapStatistics struct {
    MapName      string
    MatchesPlayed int
    
    // Statistics from STATISTICS.md
    KAST         float64  // Percentage (0-100)
    ADR          float64  // Average Damage per Round
    KD           float64  // Kill/Death ratio
    Kills        int
    Deaths       int
    FirstKills   int
    FirstDeaths  int
    TradeKills   int
    TradeDeaths  int
    
    // Additional useful stats
    Assists      int
    Headshots    int
    RoundsPlayed int
}

// WrangleResult holds the complete analysis result
type WrangleResult struct {
    PlayerStats []*PlayerStats
    MapList     []string  // Unique maps encountered
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
func ExtractPlayerFromMatch(match *api.Match, steamID64 string) *api.Player

// CalculateKAST computes KAST percentage for a player
// KAST = (Kills + Assists + Survived + Traded) / Total Rounds
func CalculateKAST(player *api.Player, match *api.Match) float64

// CalculateADR computes Average Damage per Round
func CalculateADR(player *api.Player) float64

// AggregateStats combines stats from multiple matches into OverallStats
func AggregateStats(mapStats map[string]*MapStatistics) *MapStatistics
```

### 3.4 Data Processing Pipeline

```
Input: []*api.Match + []string (SteamIDs to filter)
  ↓
1. Create PlayerStats for each SteamID
  ↓
2. For each Match:
   - Get map name
   - For each specified SteamID:
     * Find player in match (if present)
     * Extract raw stats from api.Player
     * Calculate derived stats (KAST, ADR, K/D)
     * Add to PlayerStats.MapStats[mapName]
  ↓
3. Calculate OverallStats for each player
   - Aggregate across all maps
   - Compute weighted averages for KAST, ADR
   - Sum totals for kills, deaths, etc.
  ↓
Output: WrangleResult
```

### 3.5 Statistics Calculations

**KAST (Kill, Assist, Survive, Trade):**
```
KAST = (Rounds with K + Rounds with A + Rounds survived + Rounds traded) / Total Rounds
```
This requires tracking:
- Rounds where player got a kill
- Rounds where player got an assist
- Rounds where player survived
- Rounds where player was traded (killed enemy who killed teammate recently)

**ADR (Average Damage per Round):**
```
ADR = Total Damage Dealt / Rounds Played
```

**K/D Ratio:**
```
K/D = Total Kills / Total Deaths
```
Handle division by zero (if deaths = 0, return kills as K/D)

**Trade Kills/Deaths:**
- Trade Kill: Killing an enemy within 5 seconds of them killing a teammate
- Trade Death: Being killed shortly after getting a kill
(Note: cs-demo-analyzer may provide this data, or it needs to be calculated from round events)

### 3.6 Implementation Considerations

**cs-demo-analyzer API Usage:**
The `api.Match` struct contains:
- `Players []api.Player` - Player statistics
- `Rounds []api.Round` - Round-by-round data
- `MapName string` - Map identifier

The `api.Player` struct likely contains:
- Basic info: Name, SteamID64
- Stats: Kills, Deaths, Assists, Damage, etc.
- Round-specific data may be in `api.Round` objects

**Important:** Review the cs-demo-analyzer documentation to understand:
1. What statistics are directly available
2. What needs to be calculated from raw data
3. How to extract round-by-round events for KAST and trade calculations

## 4. Display Component (GUI - Bottom Panel)

### 4.1 Table Structure

The bottom panel should display statistics in tabular format, with two view modes:

**Option A: Single Combined Table**
- Columns: Player | Map | KAST | ADR | K/D | Kills | Deaths | FK | FD | TK | TD
- Rows grouped by player, then by map
- Include overall row for each player

**Option B: Separate Tables**
- One table per player
- Ability to switch between players
- Each table shows map breakdown + overall stats

**Recommended: Option A** (simpler to implement initially)

### 4.2 Table Implementation

Using `tview.Table`:

```go
func CreateStatisticsTable(result *WrangleResult) *tview.Table {
    table := tview.NewTable().SetBorders(true).SetFixed(1, 0)
    
    // Header row
    headers := []string{"Player", "Map", "KAST%", "ADR", "K/D", 
                        "Kills", "Deaths", "FK", "FD", "TK", "TD"}
    
    // Add data rows
    // Group by player, then map
    // Include overall row for each player
    
    return table
}
```

### 4.3 Display Features

**Basic Features:**
- Sortable columns (click header to sort)
- Fixed header row
- Scrollable content
- Color coding (e.g., green for high KAST, red for low)

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

## 11. Questions for Clarification

Before implementation begins, consider clarifying:

1. **SteamID Format:** Should the application support both SteamID64 and other formats (SteamID3, SteamID32)?

2. **Player Name Usage:** Is the player name purely for display, or should it be used as a fallback if SteamID64 matching fails?

3. **Empty Player Slots:** Should all 5 player slots be mandatory, or can users analyze just 1-5 players?

4. **Demo File Extensions:** Should the app support both `.dem` and other extensions?

5. **Error Display:** Should detailed error messages (e.g., which specific demos failed) be shown in the UI or just logged?

6. **Trade Kill/Death Definition:** What exact time window constitutes a "trade" (typically 3-5 seconds)?

7. **KAST Calculation:** Does the cs-demo-analyzer library provide KAST data directly, or does it need to be calculated from round events?

8. **Table Sorting:** Should the initial sort order be by player name, overall K/D, or something else?

9. **Map Grouping:** Should maps be displayed in alphabetical order or by frequency of play?

10. **Overall Statistics:** Should "Overall" stats be weighted by rounds played per map, or simple aggregation?

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
