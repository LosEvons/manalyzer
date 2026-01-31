# Implementation Plan Audit

**Date:** 2026-01-31  
**Auditor:** AI Agent  
**Status:** In Progress

---

## Purpose

This document audits the IMPROVEMENT_PLAN.md to ensure it's accurate and sufficient for reliable AI agent implementation.

---

## Audit Criteria

1. **Technical Accuracy** - Code examples match actual codebase
2. **Completeness** - All necessary details for implementation
3. **Clarity** - Unambiguous instructions
4. **Feasibility** - Approaches are practical and tested
5. **Dependencies** - All dependencies correctly identified
6. **API Compatibility** - External libraries used correctly

---

## Issues Found

### Critical Issues (Must Fix)

#### 1. Column Index Mapping Error
**Location:** IMPROVEMENT_PLAN.md, line ~242

**Issue:**
```go
case 3: // KAST
    return players[i].OverallStats.KAST > players[j].OverallStats.KAST
case 4: // ADR
    return players[i].OverallStats.ADR > players[j].OverallStats.ADR
```

**Actual column order (from gui.go:142):**
```
0: Player
1: Map
2: Side
3: KAST%
4: ADR
5: K/D
6: Kills
7: Deaths
8: FK
9: FD
10: TK
11: TD
```

**Status:** ✅ CORRECT - Column indices are accurate

---

#### 2. Function Signature Changes
**Location:** IMPROVEMENT_PLAN.md, line ~526

**Issue:** Plan shows `createPlayerInputForm(config *Config)` but current signature is:
```go
func createPlayerInputForm() *tview.Form
```

**Impact:** HIGH - This will break during implementation

**Fix Required:** Plan must show proper migration:
1. First, modify signature to accept config parameter
2. Update callers to pass config
3. Pre-fill fields from config

**Status:** ⚠️ NEEDS CLARIFICATION

---

#### 3. UI struct modification not specified
**Location:** IMPROVEMENT_PLAN.md, line ~552

**Issue:** Code example shows `u.config.Preferences.AutoSave` but UI struct doesn't have config field

**Current UI struct:**
```go
type UI struct {
    App        *tview.Application
    Pages      *tview.Pages
    Root       *tview.Flex
    form       *tview.Form
    eventLog   *EventLog
    statsTable *StatisticsTable
}
```

**Fix Required:** Plan must specify:
```go
type UI struct {
    App        *tview.Application
    Pages      *tview.Pages
    Root       *tview.Flex
    form       *tview.Form
    eventLog   *EventLog
    statsTable *StatisticsTable
    config     *Config  // NEW: Store loaded config
}
```

**Status:** ❌ MISSING - Critical detail

---

#### 4. Missing extractConfigFromForm signature change
**Location:** IMPROVEMENT_PLAN.md, Improvement #2

**Issue:** Plan shows extracting config but doesn't show updating the return type

**Current:**
```go
func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig
```

**Needed:**
```go
func (u *UI) extractConfigFromForm(form *tview.Form) *Config
```

**Or keep separate and add:**
```go
func (u *UI) buildConfigFromForm(form *tview.Form) *Config
```

**Status:** ⚠️ AMBIGUOUS

---

### Moderate Issues (Should Fix)

#### 5. Sort implementation incomplete
**Location:** IMPROVEMENT_PLAN.md, line ~248

**Issue:**
```go
if st.sortDesc {
    // Reverse order
}
```

**Fix:** Should show actual implementation:
```go
if st.sortDesc {
    for i, j := 0, len(players)-1; i < j; i, j = i+1, j-1 {
        players[i], players[j] = players[j], players[i]
    }
}
```

**Status:** ⚠️ INCOMPLETE

---

#### 6. Missing toggleSort implementation
**Location:** IMPROVEMENT_PLAN.md, line ~229

**Issue:** Method referenced but not implemented

**Fix Required:**
```go
func (st *StatisticsTable) toggleSort(columnIndex int) {
    if st.sortColumn == columnIndex {
        // Toggle direction if same column
        st.sortDesc = !st.sortDesc
    } else {
        // New column, default to descending for stats
        st.sortColumn = columnIndex
        st.sortDesc = (columnIndex > 0) // Descending for stats, ascending for names
    }
}
```

**Status:** ❌ MISSING

---

#### 7. Header click handler may not work
**Location:** IMPROVEMENT_PLAN.md, line ~228

**Issue:**
```go
cell.SetClickedFunc(func() bool {
    st.toggleSort(columnIndex)
    st.renderTable()
    return true
})
```

**Problem:** `columnIndex` variable captured in closure may not work as expected in loop

**Fix:**
```go
for col, header := range headers {
    columnIndex := col  // Capture loop variable
    cell := tview.NewTableCell(header).
        SetTextColor(tcell.ColorYellow).
        SetAlign(tview.AlignCenter).
        SetClickedFunc(func() bool {
            st.toggleSort(columnIndex)
            st.App.QueueUpdateDraw(func() {
                st.renderTable()
            })
            return true
        })
    st.table.SetCell(0, col, cell)
}
```

**Status:** ⚠️ POTENTIAL BUG

---

#### 8. Missing StatisticsTable app reference
**Location:** IMPROVEMENT_PLAN.md, Improvement #1

**Issue:** Sort handler calls renderTable() but needs UI update

**Current struct:**
```go
type StatisticsTable struct {
    table      *tview.Table
    data       *WrangleResult
    filterMap  string
    filterSide string
}
```

**Needed:**
```go
type StatisticsTable struct {
    table      *tview.Table
    data       *WrangleResult
    filterMap  string
    filterSide string
    app        *tview.Application  // NEW: For QueueUpdateDraw
    sortColumn int
    sortDesc   bool
}
```

**Status:** ❌ MISSING

---

#### 9. go-echarts version not specified
**Location:** IMPROVEMENT_PLAN.md, line ~1026

**Issue:** Says "v2.3.3" but should verify latest stable version

**Check:**
```bash
go list -m -versions github.com/go-echarts/go-echarts/v2
```

**Status:** ⚠️ NEEDS VERIFICATION

---

#### 10. Browser auto-open error handling incomplete
**Location:** IMPROVEMENT_PLAN.md, line ~1149

**Issue:** Example shows error handling but doesn't handle "browser not found" vs "browser busy"

**Better implementation:**
```go
func openBrowser(url string) error {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "linux":
        // Try multiple browsers
        for _, browser := range []string{"xdg-open", "sensible-browser", "firefox", "chromium"} {
            if _, err := exec.LookPath(browser); err == nil {
                cmd = exec.Command(browser, url)
                break
            }
        }
        if cmd == nil {
            return fmt.Errorf("no browser found")
        }
    case "darwin":
        cmd = exec.Command("open", url)
    case "windows":
        cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
    
    return cmd.Start()
}
```

**Status:** ⚠️ INCOMPLETE

---

### Minor Issues (Nice to Fix)

#### 11. Config file permissions not specified
**Location:** IMPROVEMENT_PLAN.md, line ~457

**Issue:**
```go
return os.WriteFile(path, data, 0644)
```

**Concern:** Should config files be user-only readable?

**Recommendation:** Use `0600` for security:
```go
return os.WriteFile(path, data, 0600)
```

**Status:** ℹ️ SUGGESTION

---

#### 12. Missing visualization server shutdown
**Location:** IMPROVEMENT_PLAN.md, Improvement #4

**Issue:** HTTP server starts but no cleanup shown

**Fix Required:** Add to UI struct:
```go
type UI struct {
    // ... existing fields
    vizServer  *http.Server  // Track visualization server
}

// In shutdown/cleanup:
func (u *UI) Stop() {
    if u.vizServer != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        u.vizServer.Shutdown(ctx)
    }
    u.App.Stop()
}
```

**Status:** ℹ️ SUGGESTION

---

#### 13. Port conflict handling not detailed
**Location:** IMPROVEMENT_PLAN.md, line ~1094

**Issue:** Mentions port conflicts but doesn't show implementation

**Recommendation:**
```go
func StartVisualizationServer(data *WrangleResult) (string, error) {
    // Try ports 8080-8090
    for port := 8080; port <= 8090; port++ {
        addr := fmt.Sprintf(":%d", port)
        listener, err := net.Listen("tcp", addr)
        if err != nil {
            continue  // Port in use, try next
        }
        
        // Port available, set up server
        http.HandleFunc("/", dashboardHandler(data))
        // ... other handlers
        
        go http.Serve(listener, nil)
        return fmt.Sprintf("http://localhost:%d", port), nil
    }
    return "", fmt.Errorf("no available ports in range 8080-8090")
}
```

**Status:** ℹ️ SUGGESTION

---

#### 14. Test data generation not mentioned
**Location:** IMPROVEMENT_PLAN.md, Testing sections

**Issue:** Testing strategies mentioned but no test data generation

**Recommendation:** Add section on creating test fixtures:
```go
// test_helpers.go
func GenerateTestConfig() *Config {
    return &Config{
        Version: 1,
        Players: []PlayerConfig{
            {Name: "TestPlayer1", SteamID64: "76561198000000001"},
            {Name: "TestPlayer2", SteamID64: "76561198000000002"},
        },
        LastDemoPath: "/tmp/test/demos",
    }
}

func GenerateTestWrangleResult() *WrangleResult {
    // Generate sample player stats for testing
}
```

**Status:** ℹ️ SUGGESTION

---

## Gaps in Documentation

### 1. Migration Path for Refactoring
**What's missing:** Step-by-step guide for splitting gui.go without breaking code

**Needed:**
1. Create new files with package declaration
2. Move functions one at a time
3. Update imports
4. Verify compilation after each move
5. Test functionality after each move

---

### 2. Debugging Tips
**What's missing:** How to debug if things go wrong

**Needed:**
- How to debug tview rendering issues
- How to test config loading/saving
- How to verify filter logic
- How to test visualization server

---

### 3. Performance Considerations
**What's missing:** Performance testing and optimization

**Needed:**
- How to benchmark with large datasets
- Memory usage monitoring
- UI responsiveness testing
- Chart rendering performance

---

### 4. Error Messages
**What's missing:** What error messages to show users

**Needed:**
- Config load failure: "Failed to load configuration: %v. Using defaults."
- Config save failure: "Warning: Could not save configuration: %v"
- Visualization server failure: "Could not start visualization server: %v"
- Browser open failure: "Could not open browser. Visit http://localhost:8080"

---

## Positive Findings

### ✅ Strengths of the Plan

1. **Comprehensive research** - Multiple options evaluated for each improvement
2. **Clear rationale** - Explains why each approach was chosen
3. **Phased approach** - Logical implementation order
4. **Risk assessment** - Identifies potential issues
5. **Code examples** - Shows actual implementation
6. **Testing strategy** - Mentions testing at each phase
7. **Cross-platform** - Considers Linux, macOS, Windows
8. **Backward compatible** - Doesn't break existing features

---

## Recommendations for Fixes

### Priority 1 (Must Fix Before Implementation)

1. ✅ Add `config *Config` field to UI struct
2. ✅ Add `app *tview.Application` field to StatisticsTable
3. ✅ Add `sortColumn int` and `sortDesc bool` to StatisticsTable
4. ✅ Show function signature changes clearly
5. ✅ Complete toggleSort implementation
6. ✅ Fix header click handler closure issue
7. ✅ Complete sort reversal implementation

### Priority 2 (Should Fix)

1. ⚠️ Add proper browser auto-open with fallbacks
2. ⚠️ Add port conflict resolution
3. ⚠️ Add visualization server cleanup
4. ⚠️ Verify go-echarts version
5. ⚠️ Add migration path for refactoring

### Priority 3 (Nice to Have)

1. ℹ️ Add debugging tips
2. ℹ️ Add test data generation helpers
3. ℹ️ Add performance testing guidance
4. ℹ️ Specify user-facing error messages
5. ℹ️ Consider config file permissions

---

## Recommended Additions

### 1. Add "Implementation Checklist" Section

```markdown
## Implementation Checklist

### Before Starting
- [ ] Create feature branch
- [ ] Review current code structure
- [ ] Set up test environment
- [ ] Backup current working version

### Phase 1: Refactoring
- [ ] Create gui_form.go
- [ ] Move form-related functions
- [ ] Verify compilation
- [ ] Test UI still works
- [ ] Create gui_table.go
- [ ] Move table-related functions
- [ ] Verify compilation
- [ ] Test UI still works
- [ ] Create gui_eventlog.go
- [ ] Move event log functions
- [ ] Verify compilation
- [ ] Test UI still works
- [ ] Create config.go
- [ ] Implement LoadConfig
- [ ] Implement SaveConfig
- [ ] Add tests for config.go
- [ ] Commit refactoring

### Phase 2: Persistence
- [ ] Add config field to UI struct
- [ ] Modify New() to load config
- [ ] Modify createPlayerInputForm signature
- [ ] Pre-fill form from config
- [ ] Add config saving to onAnalyzeClicked
- [ ] Test first run (no config)
- [ ] Test save and reload
- [ ] Test manual JSON editing
- [ ] Commit persistence

### Phase 3: Filtering/Sorting
- [ ] Create filter.go
- [ ] Add sort fields to StatisticsTable
- [ ] Add app reference to StatisticsTable
- [ ] Implement toggleSort
- [ ] Implement sortData
- [ ] Add header click handlers
- [ ] Add sort indicators
- [ ] Test sorting each column
- [ ] Add filter dropdowns
- [ ] Wire dropdowns to SetFilter
- [ ] Test filter combinations
- [ ] Test filter + sort together
- [ ] Commit filtering/sorting

### Phase 4: Visualization
- [ ] Add go-echarts dependency
- [ ] Create visualize.go
- [ ] Implement HTTP server startup
- [ ] Implement browser auto-open
- [ ] Create player comparison chart
- [ ] Create T vs CT chart
- [ ] Create map breakdown chart
- [ ] Create correlation chart
- [ ] Create radar chart
- [ ] Add "Visualize" button
- [ ] Test on Linux
- [ ] Test on macOS
- [ ] Test on Windows
- [ ] Test browser not available case
- [ ] Test port conflict case
- [ ] Commit visualization
```

---

### 2. Add "Common Pitfalls" Section

```markdown
## Common Pitfalls

### Refactoring Phase
❌ **Pitfall:** Moving multiple files at once
✅ **Solution:** Move one file at a time, test between moves

❌ **Pitfall:** Forgetting to update imports
✅ **Solution:** Let compiler guide you, fix import errors one by one

### Persistence Phase
❌ **Pitfall:** Not handling missing config file
✅ **Solution:** Always check file existence, use defaults

❌ **Pitfall:** Not creating directory before writing config
✅ **Solution:** Use os.MkdirAll before os.WriteFile

### Sorting Phase
❌ **Pitfall:** Closure capturing wrong loop variable
✅ **Solution:** Create local variable: `col := col`

❌ **Pitfall:** Calling renderTable without QueueUpdateDraw
✅ **Solution:** Always wrap UI updates in QueueUpdateDraw

### Visualization Phase
❌ **Pitfall:** Forgetting HTTP server runs forever
✅ **Solution:** Store server reference, implement shutdown

❌ **Pitfall:** Assuming browser will always work
✅ **Solution:** Provide fallback (show manual URL)
```

---

### 3. Add "Verification Steps" Section

```markdown
## Verification Steps

### After Refactoring
1. Run: `go build`
2. Run: `./manalyzer`
3. Verify UI appears correctly
4. Enter test data, click Analyze
5. Verify results display correctly
6. No functionality should change

### After Persistence
1. Run: `./manalyzer` (first time)
2. Enter player data
3. Click Analyze
4. Check: Config file created at ~/.config/manalyzer/config.json
5. Exit application
6. Run: `./manalyzer` (second time)
7. Verify: Player data is pre-filled
8. Verify: Demo path is pre-filled

### After Filtering/Sorting
1. Run analysis to get data
2. Click "KAST%" header
3. Verify: Rows sorted by KAST descending
4. Verify: "▼" indicator shown
5. Click "KAST%" header again
6. Verify: Rows sorted by KAST ascending
7. Verify: "▲" indicator shown
8. Select map filter
9. Verify: Only selected map shown
10. Verify: Sorting still works

### After Visualization
1. Run analysis to get data
2. Click "Visualize" button
3. Verify: Browser opens automatically
4. Verify: Dashboard displays
5. Verify: All 5 charts render
6. Verify: Charts are interactive (hover, zoom)
7. Test: Close browser, click Visualize again
8. Verify: Works second time
```

---

## Conclusion

### Overall Assessment

**Score: 7/10** - Good foundation, needs specific fixes

**Strengths:**
- Comprehensive research ✅
- Clear recommendations ✅
- Good code examples ✅
- Logical phased approach ✅

**Weaknesses:**
- Missing critical struct field additions ❌
- Some incomplete code examples ⚠️
- Function signature changes not clearly shown ⚠️
- Closure bug in sorting implementation ⚠️

### Is It Ready for AI Implementation?

**Status: NOT YET** 🔴

**Why not:**
1. Missing critical struct field additions (UI.config, StatisticsTable.app, etc.)
2. Function signature changes would cause compilation errors
3. Some code examples are incomplete
4. Closure bug would cause runtime issues

**What's needed:**
1. Fix all Priority 1 issues (struct fields, signatures, implementations)
2. Add implementation checklist
3. Add common pitfalls section
4. Add verification steps

**After fixes:** Should be ready for AI implementation ✅

---

## Next Actions

1. ✅ Create this audit document
2. ⬜ Fix all Priority 1 issues in IMPROVEMENT_PLAN.md
3. ⬜ Add implementation checklist
4. ⬜ Add common pitfalls section
5. ⬜ Add verification steps section
6. ⬜ Re-review plan after fixes
7. ⬜ Mark as "Ready for Implementation"

---

**End of Audit**
