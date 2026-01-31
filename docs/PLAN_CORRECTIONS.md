# Implementation Plan - Corrections and Additions

**Date:** 2026-01-31  
**Purpose:** Critical fixes and additions to make IMPROVEMENT_PLAN.md AI-agent-ready

---

## SECTION 1: Struct Field Additions (CRITICAL)

### Problem
The plan references struct fields that don't exist in current code.

### Fix: Update UI Struct

**File:** `src/gui.go` (after refactoring: still `gui.go`)

**Add config field:**
```go
type UI struct {
    App        *tview.Application
    Pages      *tview.Pages
    Root       *tview.Flex
    form       *tview.Form
    eventLog   *EventLog
    statsTable *StatisticsTable
    config     *Config  // NEW: Store loaded configuration
}
```

### Fix: Update StatisticsTable Struct

**File:** `src/gui_table.go` (after refactoring)

**Add sorting and app fields:**
```go
type StatisticsTable struct {
    table      *tview.Table
    data       *WrangleResult
    filterMap  string
    filterSide string
    app        *tview.Application  // NEW: For UI updates
    sortColumn int                 // NEW: Current sort column (0-11)
    sortDesc   bool                // NEW: true = descending
}
```

---

## SECTION 2: Function Signatures (CRITICAL)

### Problem
Plan shows function calls that would break compilation.

### Fix 1: createPlayerInputForm

**Step 1:** Keep current signature, add new function

```go
// Keep existing (for backward compatibility during refactoring)
func createPlayerInputForm() *tview.Form {
    return createPlayerInputFormWithConfig(DefaultConfig())
}

// New function that accepts config
func createPlayerInputFormWithConfig(config *Config) *tview.Form {
    form := tview.NewForm()
    
    form.SetBorder(true)
    form.SetTitle("Player Configuration")
    form.SetTitleAlign(tview.AlignLeft)

    // Add player inputs (pre-filled from config)
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

    // Add demo path (pre-filled from config)
    form.AddInputField("Demo Base Path", config.LastDemoPath, 50, nil, nil)

    // Add buttons
    form.AddButton("Analyze", nil) // Handler added later
    form.AddButton("Clear", nil)

    return form
}
```

**Step 2:** Update New() function

```go
func New() *UI {
    app := tview.NewApplication()

    // Load config
    config, err := LoadConfig()
    if err != nil {
        log.Printf("Failed to load config: %v, using defaults", err)
        config = DefaultConfig()
    }

    // Create components with config
    form := createPlayerInputFormWithConfig(config)
    eventLog := newEventLog(50)
    statsTable := newStatisticsTable(app)  // Pass app reference

    // ... rest of layout setup ...

    ui := &UI{
        App:        app,
        Pages:      pages,
        Root:       mainLayout,
        form:       form,
        eventLog:   eventLog,
        statsTable: statsTable,
        config:     config,  // Store config
    }

    // Setup handlers after UI is created
    ui.setupFormHandlers(form)

    return ui
}
```

### Fix 2: newStatisticsTable

**Update signature to accept app:**

```go
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
```

### Fix 3: extractConfigFromForm

**Keep existing, add new function:**

```go
// Existing - returns AnalysisConfig (temporary data for analysis)
func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig {
    // ... existing implementation ...
}

// New - builds Config for persistence
func (u *UI) buildConfigFromForm(form *tview.Form) *Config {
    config := &Config{
        Version: 1,
        Players: make([]PlayerConfig, 0),
        Preferences: Preferences{
            AutoSave: true,
        },
    }
    
    // Extract player data (5 pairs of name + steamID)
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
        
        // Only add if SteamID is not empty
        if steamID != "" {
            config.Players = append(config.Players, PlayerConfig{
                Name:      name,
                SteamID64: steamID,
            })
        }
    }

    // Extract base path (index 10 = after 5 player pairs)
    if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
        config.LastDemoPath = pathField.GetText()
    }

    return config
}
```

---

## SECTION 3: Complete Implementations (CRITICAL)

### Implementation 1: toggleSort

```go
func (st *StatisticsTable) toggleSort(columnIndex int) {
    if st.sortColumn == columnIndex {
        // Same column - toggle direction
        st.sortDesc = !st.sortDesc
    } else {
        // New column - set default direction
        st.sortColumn = columnIndex
        
        // Default descending for stats columns, ascending for names
        if columnIndex == 0 || columnIndex == 1 {
            // Player name or Map name - ascending by default
            st.sortDesc = false
        } else {
            // Stats - descending by default (higher is better)
            st.sortDesc = true
        }
    }
}
```

### Implementation 2: sortData (Complete)

```go
func (st *StatisticsTable) sortData(players []*PlayerStats) []*PlayerStats {
    // Make a copy to avoid modifying original
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
        case 1: // Map - not sortable (multiple maps per player)
            less = sorted[i].PlayerName < sorted[j].PlayerName
        case 2: // Side - not sortable
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
        case 8: // First Kills
            if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
                less = sorted[i].OverallStats.FirstKills < sorted[j].OverallStats.FirstKills
            }
        case 9: // First Deaths
            if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
                less = sorted[i].OverallStats.FirstDeaths < sorted[j].OverallStats.FirstDeaths
            }
        case 10: // Trade Kills
            if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
                less = sorted[i].OverallStats.TradeKills < sorted[j].OverallStats.TradeKills
            }
        case 11: // Trade Deaths
            if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
                less = sorted[i].OverallStats.TradeDeaths < sorted[j].OverallStats.TradeDeaths
            }
        default:
            less = sorted[i].PlayerName < sorted[j].PlayerName
        }
        
        // Apply sort direction
        if st.sortDesc {
            return !less
        }
        return less
    })
    
    return sorted
}
```

### Implementation 3: renderTable with Sorting

**In renderTable(), replace line 158-170 with:**

```go
// Sort players
sortedPlayers := st.sortData(playerStatsSlice)

for _, playerStats := range sortedPlayers {
    if playerStats == nil {
        continue
    }
    
    // ... rest of rendering logic ...
}
```

### Implementation 4: Header Click Handlers (Fixed)

```go
func (st *StatisticsTable) renderTable() {
    st.table.Clear()

    // Header row with column names
    headers := []string{"Player", "Map", "Side", "KAST%", "ADR", "K/D",
        "Kills", "Deaths", "FK", "FD", "TK", "TD"}

    for col, header := range headers {
        columnIndex := col  // Capture loop variable
        
        // Add sort indicator if this is the sorted column
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
        
        // Make stats columns clickable for sorting
        if columnIndex >= 3 {  // Only stat columns are sortable
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

    // Rest of rendering logic...
}
```

---

## SECTION 4: Config Save Integration (CRITICAL)

### Update onAnalyzeClicked

```go
func (u *UI) onAnalyzeClicked(form *tview.Form) {
    // Collect form data
    analysisConfig := u.extractConfigFromForm(form)

    // Validate base path first
    if analysisConfig.BasePath == "" {
        u.logEvent("Error: Demo base path must be specified")
        return
    }

    if _, err := os.Stat(analysisConfig.BasePath); os.IsNotExist(err) {
        u.logEvent(fmt.Sprintf("Error: Path does not exist: %s", analysisConfig.BasePath))
        return
    }

    // Validate at least one player is specified
    validPlayers := 0
    for _, player := range analysisConfig.Players {
        if player.SteamID64 != "" {
            validPlayers++
        }
    }

    if validPlayers == 0 {
        u.logEvent("Error: At least one player with SteamID64 must be specified")
        return
    }

    // Save configuration (if auto-save enabled)
    if u.config.Preferences.AutoSave {
        persistConfig := u.buildConfigFromForm(form)
        if err := SaveConfig(persistConfig); err != nil {
            u.logEvent(fmt.Sprintf("Warning: Failed to save config: %v", err))
        } else {
            u.config = persistConfig  // Update cached config
        }
    }

    // Start analysis (in goroutine to keep UI responsive)
    go u.runAnalysis(analysisConfig)
}
```

---

## SECTION 5: Browser Auto-Open (IMPROVED)

```go
// openBrowser opens a URL in the system's default browser.
// It tries multiple methods and browsers depending on the OS.
func openBrowser(url string) error {
    var cmd *exec.Cmd
    var err error
    
    switch runtime.GOOS {
    case "linux":
        // Try multiple browsers in order of preference
        browsers := []string{"xdg-open", "sensible-browser", "firefox", "chromium-browser", "google-chrome"}
        for _, browser := range browsers {
            if _, err := exec.LookPath(browser); err == nil {
                cmd = exec.Command(browser, url)
                return cmd.Start()
            }
        }
        return fmt.Errorf("no browser found (tried: %s)", strings.Join(browsers, ", "))
        
    case "darwin": // macOS
        cmd = exec.Command("open", url)
        return cmd.Start()
        
    case "windows":
        cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
        return cmd.Start()
        
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}
```

---

## SECTION 6: Visualization Server with Port Handling

```go
// StartVisualizationServer starts the HTTP server for visualization.
// Returns the URL the server is listening on, or an error.
func StartVisualizationServer(data *WrangleResult) (string, error) {
    // Try to find an available port
    var listener net.Listener
    var port int
    var err error
    
    for port = 8080; port <= 8090; port++ {
        addr := fmt.Sprintf(":%d", port)
        listener, err = net.Listen("tcp", addr)
        if err == nil {
            break  // Found available port
        }
    }
    
    if err != nil {
        return "", fmt.Errorf("no available ports in range 8080-8090")
    }
    
    // Set up handlers
    mux := http.NewServeMux()
    mux.HandleFunc("/", dashboardHandler(data))
    mux.HandleFunc("/player-comparison", playerComparisonHandler(data))
    mux.HandleFunc("/map-breakdown", mapBreakdownHandler(data))
    mux.HandleFunc("/side-performance", sidePerformanceHandler(data))
    mux.HandleFunc("/stat-correlation", statCorrelationHandler(data))
    
    // Start server in background
    server := &http.Server{Handler: mux}
    go server.Serve(listener)
    
    url := fmt.Sprintf("http://localhost:%d", port)
    return url, nil
}
```

---

## SECTION 7: UI Integration for Visualization

```go
// In setupFormHandlers, add Visualize button
func (u *UI) setupFormHandlers(form *tview.Form) {
    // Get button indices
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
    
    // Add Visualize button
    form.AddButton("Visualize", func() {
        u.onVisualizeClicked()
    })
}

func (u *UI) onVisualizeClicked() {
    if u.statsTable.data == nil {
        u.logEvent("Error: No data to visualize. Run analysis first.")
        return
    }
    
    u.logEvent("Starting visualization server...")
    
    // Start server and get URL
    url, err := StartVisualizationServer(u.statsTable.data)
    if err != nil {
        u.logEvent(fmt.Sprintf("Error: Failed to start visualization server: %v", err))
        return
    }
    
    u.logEvent(fmt.Sprintf("Visualization server started at %s", url))
    
    // Try to open browser
    time.Sleep(500 * time.Millisecond)  // Give server time to start
    
    if err := openBrowser(url); err != nil {
        u.logEvent(fmt.Sprintf("Could not open browser: %v", err))
        u.logEvent(fmt.Sprintf("Visit %s manually to view charts", url))
    } else {
        u.logEvent("Visualization dashboard opened in browser")
    }
}
```

---

## SECTION 8: Implementation Checklist

Add this to IMPROVEMENT_PLAN.md after Risk Assessment section:

```markdown
## Implementation Checklist

Use this checklist to track progress during implementation.

### Pre-Implementation
- [ ] Create feature branch: `git checkout -b feature/improvements`
- [ ] Review current code structure
- [ ] Set up test environment with sample .dem files
- [ ] Document current behavior for regression testing

### Phase 1: Refactoring (Week 1, Day 1-2)

#### Day 1: Split GUI Files
- [ ] Create `src/gui_eventlog.go`
- [ ] Move EventLog struct and methods
- [ ] Add package declaration and imports
- [ ] Compile: `go build` - fix any errors
- [ ] Test: Run app, verify event log works
- [ ] Commit: "refactor: extract EventLog to gui_eventlog.go"

- [ ] Create `src/gui_form.go`
- [ ] Move form-related functions (createPlayerInputForm, validateSteamID64, setupFormHandlers, etc.)
- [ ] Add package declaration and imports
- [ ] Compile: `go build` - fix any errors
- [ ] Test: Run app, verify form works
- [ ] Commit: "refactor: extract form functions to gui_form.go"

#### Day 2: Complete Refactoring
- [ ] Create `src/gui_table.go`
- [ ] Move StatisticsTable struct and all methods
- [ ] Add package declaration and imports
- [ ] Compile: `go build` - fix any errors
- [ ] Test: Run app with analysis, verify table rendering
- [ ] Commit: "refactor: extract StatisticsTable to gui_table.go"

- [ ] Update `src/gui.go`
- [ ] Keep only UI struct, New(), Start(), Stop(), QueueUpdate(), logEvent()
- [ ] Verify all other functions moved
- [ ] Compile: `go build` - should be clean
- [ ] Test: Full app functionality
- [ ] Commit: "refactor: clean up gui.go, keep only core UI functions"

- [ ] Create `src/config.go`
- [ ] Implement Config struct
- [ ] Implement LoadConfig()
- [ ] Implement SaveConfig()
- [ ] Implement DefaultConfig()
- [ ] Implement configPath()
- [ ] Add package declaration
- [ ] Compile: `go build`
- [ ] Commit: "feat: add config.go with persistence structures"

### Phase 2: Persistent Storage (Week 1, Day 3-4)

#### Day 3: Config Integration
- [ ] Add `config *Config` field to UI struct
- [ ] Modify New() to load config on startup
- [ ] Update createPlayerInputForm to createPlayerInputFormWithConfig
- [ ] Pre-fill form fields from loaded config
- [ ] Compile: `go build`
- [ ] Test: First run (no config file) - should use defaults
- [ ] Commit: "feat: load config on startup and pre-fill form"

#### Day 4: Config Saving
- [ ] Add buildConfigFromForm() method
- [ ] Update onAnalyzeClicked to save config
- [ ] Add "Save Config" button (optional)
- [ ] Test: Enter data, click Analyze, check config file created
- [ ] Test: Restart app, verify data loaded
- [ ] Test: Manually edit JSON, verify changes loaded
- [ ] Test: Corrupt JSON, verify defaults used
- [ ] Commit: "feat: save config on analyze"

- [ ] Update README.md with config file location
- [ ] Test on Linux: ~/.config/manalyzer/config.json
- [ ] Test on macOS: ~/Library/Application Support/manalyzer/config.json (or ~/.config/)
- [ ] Test on Windows: %APPDATA%\manalyzer\config.json (or %USERPROFILE%\.config\)
- [ ] Commit: "docs: document config file location"

### Phase 3: Filtering and Sorting (Week 2, Day 1-3)

#### Day 1: Sorting Implementation
- [ ] Create `src/filter.go` (empty for now, for future use)
- [ ] Add sortColumn and sortDesc fields to StatisticsTable
- [ ] Add app field to StatisticsTable
- [ ] Update newStatisticsTable to accept app parameter
- [ ] Update New() to pass app to newStatisticsTable
- [ ] Compile: `go build`
- [ ] Commit: "feat: add sort state to StatisticsTable"

- [ ] Implement toggleSort() method
- [ ] Implement sortData() method
- [ ] Update renderTable() to use sortData()
- [ ] Compile: `go build`
- [ ] Test: Should still work (default sort by name)
- [ ] Commit: "feat: implement sorting logic"

#### Day 2: Interactive Sorting
- [ ] Update renderTable() to add sort indicators (▲/▼)
- [ ] Add click handlers to header cells
- [ ] Fix closure bug (capture columnIndex)
- [ ] Wrap renderTable() call in QueueUpdateDraw
- [ ] Compile: `go build`
- [ ] Test: Click each column header
- [ ] Test: Verify sort direction toggles
- [ ] Test: Verify indicator updates
- [ ] Commit: "feat: add clickable column headers for sorting"

#### Day 3: Filter UI
- [ ] Add dropdown for Map filter (above table)
- [ ] Add dropdown for Side filter (T/CT/All)
- [ ] Wire dropdowns to SetFilter()
- [ ] Test: Select map, verify filtering
- [ ] Test: Select side, verify filtering
- [ ] Test: Combine filters, verify both work
- [ ] Test: Sort with filters active
- [ ] Commit: "feat: add filter dropdowns for map and side"

### Phase 4: Data Visualization (Week 3, Day 1-5)

#### Day 1: Infrastructure
- [ ] Add go-echarts dependency: `go get github.com/go-echarts/go-echarts/v2`
- [ ] Create `src/visualize.go`
- [ ] Implement StartVisualizationServer with port handling
- [ ] Implement openBrowser with fallbacks
- [ ] Add basic dashboard handler (placeholder HTML)
- [ ] Compile: `go build`
- [ ] Test: Call StartVisualizationServer manually
- [ ] Test: Verify browser opens
- [ ] Commit: "feat: add visualization server infrastructure"

#### Day 2: Chart Implementation (Part 1)
- [ ] Implement playerComparisonHandler (bar chart)
- [ ] Test: Access http://localhost:8080/player-comparison
- [ ] Verify: Chart displays correctly
- [ ] Implement sidePerformanceHandler (grouped bar chart)
- [ ] Test: Access http://localhost:8080/side-performance
- [ ] Verify: Chart displays correctly
- [ ] Commit: "feat: implement player comparison and side performance charts"

#### Day 3: Chart Implementation (Part 2)
- [ ] Implement mapBreakdownHandler (heatmap)
- [ ] Test: Access http://localhost:8080/map-breakdown
- [ ] Verify: Chart displays correctly
- [ ] Implement statCorrelationHandler (scatter plot)
- [ ] Test: Access http://localhost:8080/stat-correlation
- [ ] Verify: Chart displays correctly
- [ ] Commit: "feat: implement map breakdown and stat correlation charts"

#### Day 4: Dashboard and Integration
- [ ] Implement dashboardHandler (main page with all charts)
- [ ] Add CSS styling to dashboard
- [ ] Test: Access http://localhost:8080/
- [ ] Verify: All charts display on one page
- [ ] Add "Visualize" button to GUI
- [ ] Implement onVisualizeClicked
- [ ] Test: Click button, verify dashboard opens
- [ ] Commit: "feat: add dashboard and integrate with GUI"

#### Day 5: Testing and Polish
- [ ] Test: No data (should show error message)
- [ ] Test: Port 8080 in use (should try 8081-8090)
- [ ] Test: Browser not available (should show manual URL)
- [ ] Test: Multiple visualize clicks (should work)
- [ ] Test on Linux
- [ ] Test on macOS
- [ ] Test on Windows
- [ ] Fix any platform-specific issues
- [ ] Commit: "test: verify visualization on all platforms"

### Final Steps

#### Documentation
- [ ] Update README.md with new features
- [ ] Add screenshots of visualization
- [ ] Document keyboard shortcuts
- [ ] Document config file format
- [ ] Update usage guide
- [ ] Commit: "docs: update documentation for new features"

#### Testing
- [ ] Run full regression test
- [ ] Test with 1 player
- [ ] Test with 5 players
- [ ] Test with large dataset (100+ demos)
- [ ] Test with corrupt demo files
- [ ] Verify performance is acceptable
- [ ] Fix any issues found

#### Release
- [ ] Create PR with all commits
- [ ] Request review
- [ ] Address feedback
- [ ] Merge to main
- [ ] Tag release: v1.1.0
- [ ] Create release notes
```

---

## SECTION 9: Common Pitfalls

Add this section to IMPROVEMENT_PLAN.md:

```markdown
## Common Pitfalls and Solutions

### Refactoring Phase

❌ **Pitfall:** Moving multiple functions at once and getting import errors  
✅ **Solution:** Move one struct/function at a time, compile after each move

❌ **Pitfall:** Forgetting package declaration in new files  
✅ **Solution:** Every .go file must start with `package manalyzer`

❌ **Pitfall:** Circular imports between gui files  
✅ **Solution:** All gui files are in same package, no imports between them needed

### Persistence Phase

❌ **Pitfall:** Config file not created (directory doesn't exist)  
✅ **Solution:** Use `os.MkdirAll(dir, 0755)` before `os.WriteFile()`

❌ **Pitfall:** App crashes on first run (config file missing)  
✅ **Solution:** Always check if file exists with `os.Stat()`, use defaults if not

❌ **Pitfall:** Config saved but changes not visible on restart  
✅ **Solution:** Verify config.Players slice is populated, not empty

### Sorting Phase

❌ **Pitfall:** Click handler doesn't work (closure captures wrong variable)  
✅ **Solution:** Create local variable inside loop: `columnIndex := col`

❌ **Pitfall:** Sort works but UI doesn't update  
✅ **Solution:** Wrap renderTable() in `app.QueueUpdateDraw(func() { ... })`

❌ **Pitfall:** Panic: nil pointer dereference in sort  
✅ **Solution:** Check `if player.OverallStats != nil` before accessing fields

❌ **Pitfall:** Sort indicator doesn't appear  
✅ **Solution:** Append " ▼" or " ▲" to header string before creating cell

### Visualization Phase

❌ **Pitfall:** HTTP server blocks the UI  
✅ **Solution:** Start server in goroutine: `go http.ListenAndServe(...)`

❌ **Pitfall:** Port 8080 already in use  
✅ **Solution:** Try ports 8080-8090, use first available

❌ **Pitfall:** Browser doesn't open  
✅ **Solution:** Provide fallback message with manual URL

❌ **Pitfall:** Charts don't render (JavaScript errors)  
✅ **Solution:** Check browser console, verify go-echarts HTML generation

❌ **Pitfall:** Server keeps running after app exits  
✅ **Solution:** Store server reference, call Shutdown() in Stop()
```

---

## SECTION 10: Verification Steps

Add this section to IMPROVEMENT_PLAN.md:

```markdown
## Verification Steps

### After Each Phase

#### After Refactoring
1. ✅ Run: `go build` - should compile without errors
2. ✅ Run: `./manalyzer` - app should start
3. ✅ Verify: UI looks identical to before
4. ✅ Test: Enter player data, click Analyze
5. ✅ Verify: Results display correctly
6. ✅ Test: Click Clear button
7. ✅ Verify: Form clears
8. **Expected:** No functionality changes at all

#### After Persistence
1. ✅ Delete any existing config file
2. ✅ Run: `./manalyzer` (first time)
3. ✅ Verify: Form is empty (default values)
4. ✅ Enter: Player names and SteamID64s
5. ✅ Enter: Demo path
6. ✅ Click: Analyze button
7. ✅ Check: Config file created at ~/.config/manalyzer/config.json
8. ✅ Exit: Application (Ctrl+C or ESC)
9. ✅ Run: `./manalyzer` (second time)
10. ✅ Verify: Form is pre-filled with previous data
11. ✅ Test: Change one field, click Analyze
12. ✅ Restart: Application
13. ✅ Verify: Changed field is persisted
14. ✅ Test: Manually edit JSON file (change a name)
15. ✅ Restart: Application
16. ✅ Verify: Manual changes are loaded

#### After Filtering/Sorting
1. ✅ Run: `./manalyzer` with test data
2. ✅ Click: Analyze to get results
3. ✅ Verify: Table displays with data
4. ✅ Click: "KAST%" header
5. ✅ Verify: Rows sorted by KAST descending
6. ✅ Verify: "▼" indicator shown in KAST% header
7. ✅ Click: "KAST%" header again
8. ✅ Verify: Rows sorted by KAST ascending
9. ✅ Verify: "▲" indicator shown in KAST% header
10. ✅ Click: "ADR" header
11. ✅ Verify: Rows sorted by ADR descending
12. ✅ Verify: "▼" indicator moved to ADR header
13. ✅ Select: Map filter (choose a specific map)
14. ✅ Verify: Only rows for selected map shown
15. ✅ Verify: Sorting still works
16. ✅ Select: Side filter (choose T or CT)
17. ✅ Verify: Only rows for selected side shown
18. ✅ Verify: Sorting still works
19. ✅ Reset: Filters to "All"
20. ✅ Verify: All data shown again

#### After Visualization
1. ✅ Run: `./manalyzer` with test data
2. ✅ Click: Analyze to get results
3. ✅ Click: "Visualize" button
4. ✅ Verify: Event log shows "Starting visualization server..."
5. ✅ Verify: Event log shows URL (http://localhost:808X)
6. ✅ Verify: Browser opens automatically
7. ✅ Verify: Dashboard page loads
8. ✅ Verify: All 5 charts are visible:
   - Player comparison bar chart
   - T vs CT grouped bar chart
   - Map breakdown heatmap
   - Stat correlation scatter plot
   - Player radar chart
9. ✅ Test: Hover over chart elements
10. ✅ Verify: Tooltips appear with data
11. ✅ Test: Zoom in/out on charts (if supported)
12. ✅ Verify: Charts are interactive
13. ✅ Close: Browser
14. ✅ Click: "Visualize" button again
15. ✅ Verify: Works second time (reuses server or starts new one)
16. ✅ Test: Kill server process externally
17. ✅ Click: "Visualize" button
18. ✅ Verify: New server starts successfully
19. ✅ Test: Change port 8080 to in-use (e.g., `nc -l 8080`)
20. ✅ Click: "Visualize" button
21. ✅ Verify: Server starts on 8081 or next available port
22. ✅ Test: Disable browser auto-open (modify openBrowser to return error)
23. ✅ Click: "Visualize" button
24. ✅ Verify: Manual URL shown in event log
```

---

## Summary of Required Changes to IMPROVEMENT_PLAN.md

### Add These Sections:
1. ✅ Struct field additions (UI.config, StatisticsTable.app/sortColumn/sortDesc)
2. ✅ Complete function implementations (toggleSort, sortData, fixed header handlers)
3. ✅ Browser auto-open with fallbacks
4. ✅ Visualization server with port handling
5. ✅ Implementation checklist (detailed day-by-day)
6. ✅ Common pitfalls section
7. ✅ Verification steps section

### Update These Sections:
1. ⚠️ Fix function signature examples (show migration path)
2. ⚠️ Fix closure bug in sorting example
3. ⚠️ Complete sort reversal implementation
4. ⚠️ Show proper config save integration

### Verify These Details:
1. ✅ Column indices (already correct)
2. ⚠️ go-echarts version (verify latest stable)
3. ✅ Config file permissions (recommend 0600)

---

**END OF CORRECTIONS**
