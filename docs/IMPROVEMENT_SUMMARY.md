# Manalyzer Improvements - Executive Summary

**Date:** 2026-01-31  
**Status:** Research Complete - Awaiting Approval  
**Full Plan:** See [IMPROVEMENT_PLAN.md](./IMPROVEMENT_PLAN.md)

---

## Quick Overview

This document summarizes research for four major improvements to Manalyzer. Each improvement has been thoroughly researched with specific recommendations.

---

## 1. Filtering, Sorting, and Grouping ⭐ High Priority

### The Problem
- Users can't filter table data by map or side (logic exists but no UI)
- Can't sort by any column except player name
- Large datasets are hard to navigate

### The Solution
**Add interactive table controls:**
- ✅ Dropdown filters above table (Map, Side, Player)
- ✅ Clickable column headers for sorting
- ✅ Sort indicators (▲/▼)
- ⚠️ Optional: Stat range filters (KAST > 70%, etc.)

### Implementation
- **Effort:** Medium (6-9 hours)
- **Tech:** tview dropdowns + custom sort logic
- **Dependencies:** None (existing tview)
- **Risk:** Low

---

## 2. Persistent Storage ⭐ High Priority

### The Problem
- Users must re-enter 5 players (name + SteamID64) every session
- Demo path not remembered
- Very tedious for regular users

### The Solution
**JSON config file with auto-save/load:**

```json
{
  "players": [
    {"name": "Player1", "steamID64": "76561198012345678"},
    {"name": "Player2", "steamID64": "76561198087654321"}
  ],
  "lastDemoPath": "/path/to/demos"
}
```

**Location:** `~/.config/manalyzer/config.json` (Linux/Mac) or `%APPDATA%\manalyzer\config.json` (Windows)

**Behavior:**
- Auto-load on startup (pre-fill form)
- Auto-save when "Analyze" clicked
- Optional "Save Config" button

### Implementation
- **Effort:** Small (3-4 hours)
- **Tech:** Standard library `encoding/json`
- **Dependencies:** None
- **Risk:** Very Low

---

## 3. Refactoring ⭐ High Priority

### The Problem
- `gui.go` is 605 lines (too large, multiple responsibilities)
- Some functions in `wrangle.go` exceed 100 lines
- Hard to navigate and maintain

### The Solution
**Split into focused files:**

```
Before:                 After:
src/                    src/
├── gui.go (605 lines) ├── gui.go (~100 lines)         # Main UI
├── wrangle.go         ├── gui_form.go (~150 lines)    # Form & handlers
├── gather.go          ├── gui_table.go (~250 lines)   # Statistics table
└── main.go            ├── gui_eventlog.go (~80 lines) # Event log
                       ├── config.go (~150 lines)      # Config management
                       ├── filter.go (~100 lines)      # Filtering/sorting
                       ├── wrangle.go (~400 lines)     # Stats processing
                       └── gather.go (~126 lines)      # Demo gathering
```

**Benefits:**
- Each file <300 lines
- Clear separation of concerns
- Easier to find and modify code
- Better for future features

### Implementation
- **Effort:** Small (4-6 hours)
- **Tech:** Just file reorganization
- **Dependencies:** None
- **Risk:** Very Low (mechanical refactoring)

---

## 4. Data Visualization 🎨 Medium Priority

### The Problem
- No way to visually compare player performance
- Hard to spot patterns or trends in table data
- Need charts for effective analysis

### The Research Question
**"What is the best way to implement data visualization?"**

### Options Evaluated

| Option | Description | Verdict |
|--------|-------------|---------|
| **Terminal ASCII charts** | termui, asciigraph | ❌ Conflicts with tview, limited |
| **Web dashboard** | go-echarts + HTTP server | ✅ **RECOMMENDED** |
| **Static image export** | gonum/plot → PNG files | ⚠️ Fallback if web rejected |
| **No charts** | Keep table-only | ❌ Doesn't meet requirements |

### The Recommendation: Hybrid Approach 🏆

**Keep terminal UI + Add web dashboard**

```
┌─────────────────────┐
│  Terminal UI        │  ← Keep existing (primary interface)
│  (tview)            │
│                     │
│  [Analyze]          │
│  [Visualize] ← NEW  │  Click to open charts in browser
└─────────────────────┘
         │
         │ Opens browser
         ▼
┌─────────────────────┐
│  Web Dashboard      │  ← New feature (optional)
│  Interactive Charts │
│  - Bar charts       │
│  - Heatmaps         │
│  - Scatter plots    │
│  - Radar charts     │
└─────────────────────┘
```

**Why this approach?**
- ✅ Best of both worlds (terminal + rich visuals)
- ✅ go-echarts is pure Go, mature, well-maintained
- ✅ Interactive charts (zoom, hover, export)
- ✅ No conflicts with tview (separate HTTP server)
- ✅ Charts work alongside terminal (dual-interface)
- ✅ Optional - terminal works fine without it

**Architecture:**
1. User clicks "Visualize" button in terminal UI
2. App starts HTTP server on localhost:8080
3. Browser auto-opens with dashboard
4. User can view 5 chart types:
   - Player comparison (bar chart)
   - T vs CT performance (grouped bar)
   - Map breakdown (heatmap)
   - Stat correlation (scatter plot)
   - Player profile (radar chart)

**Example Chart Types:**

```
Player Comparison                 T vs CT Performance
───────────────────              ───────────────────
KAST%  ████████ P1 (75.0)       Player1  ██ T  ███ CT
       █████ P2 (60.0)          Player2  ███ T ██ CT
       ███████ P3 (70.0)        Player3  ████ T ████ CT

Map Breakdown (Heatmap)          Stat Correlation
───────────────────              ───────────────────
        dust2  mirage cache           K/D ▲
Player1  🟢     🟡     🔴              │  o  o
Player2  🟡     🟢     🟢              │    o
Player3  🔴     🟡     🟡              └──────► KAST%
```

### Implementation
- **Effort:** Medium (6-8 hours)
- **Tech:** go-echarts library (pure Go)
- **Dependencies:** `github.com/go-echarts/go-echarts/v2`
- **Risk:** Low-Medium
  - Port conflicts (mitigated: dynamic port selection)
  - Browser not available (mitigated: show manual URL)

### Alternative If Web Is Rejected
- Use gonum/plot to export PNG images
- Less interactive but still useful
- No browser dependency

---

## Implementation Roadmap

### Recommended Order

```
Phase 1 (Week 1): Foundation
├── Refactoring (#3)      ← Clean up code first
└── Persistence (#2)      ← Easy win, high value

Phase 2 (Week 2): Table Enhancements
└── Filtering/Sorting (#1) ← Build on clean code

Phase 3 (Week 3-4): Visualization
└── Web Dashboard (#4)    ← Final feature add
```

### Timeline

| Phase | Duration | Features | Deliverables |
|-------|----------|----------|--------------|
| **Phase 1** | Week 1 (7-10 hrs) | Refactoring + Config | Clean code + persistence |
| **Phase 2** | Week 2 (6-9 hrs) | Filters + Sorting | Enhanced table interaction |
| **Phase 3** | Week 3-4 (6-8 hrs) | Visualization | Interactive charts |
| **Total** | **3-4 weeks** | **All 4 improvements** | **Production-ready** |

---

## Effort Summary

| Improvement | Effort | Priority | Dependencies |
|-------------|--------|----------|--------------|
| **#1 Filtering/Sorting** | 6-9 hours | High | None |
| **#2 Persistence** | 3-4 hours | High | None (stdlib) |
| **#3 Refactoring** | 4-6 hours | High | None |
| **#4 Visualization** | 6-8 hours | Medium | go-echarts |
| **Total (Core)** | **19-27 hours** | - | **1 new dependency** |
| **Total (w/ Optional)** | **27-36 hours** | - | - |

---

## Risk Assessment

| Risk Category | Level | Mitigation |
|---------------|-------|------------|
| **Breaking existing features** | Low | Incremental testing, refactor first |
| **Performance with large datasets** | Low | Test with 100+ matches |
| **Cross-platform issues** | Low | Test on Linux, macOS, Windows |
| **Browser not available** | Low | Provide manual URL in terminal |
| **Scope creep** | Medium | Stick to defined phases, defer extras |

---

## Key Decisions Required

### 1. Visualization Approach
**Options:**
- A. Hybrid (Terminal + Web Dashboard with go-echarts) ← **RECOMMENDED**
- B. Static PNG export (gonum/plot)
- C. Skip visualization entirely
- D. Terminal ASCII charts (termui) ← Not recommended

**Recommendation:** Choose option A (Hybrid)

### 2. Optional Features
Which optional features to include?
- [ ] Stat range filters (e.g., KAST > 70%)
- [ ] "Save Config" button (beyond auto-save)
- [ ] Chart export to PNG/SVG
- [ ] Alternative table grouping views

**Recommendation:** Start with core features, add optional based on time/feedback

### 3. Implementation Priority
All improvements or subset?
- **All 4** improvements (19-27 hours)
- **Top 3** (exclude visualization) (13-19 hours)
- **Top 2** (refactoring + persistence) (7-10 hours)

**Recommendation:** All 4 improvements for complete solution

---

## Success Criteria

### Must-Have (MVP)

- [x] Research complete ✅
- [ ] All improvements have clear approach
- [ ] User approves plan
- [ ] Implementation begins

**After Implementation:**

- [ ] Config persists between sessions
- [ ] Users can filter and sort table data
- [ ] Code is clean and maintainable (<300 lines per file)
- [ ] Users can view interactive charts (if #4 approved)
- [ ] All existing features still work
- [ ] Tested on Linux, macOS, Windows

### Nice-to-Have

- [ ] Stat range filters
- [ ] Chart export functionality
- [ ] Alternative grouping views
- [ ] "Reset Config" button

---

## Next Steps

### For User Review

1. **Read this summary**
2. **Review detailed plan** at [IMPROVEMENT_PLAN.md](./IMPROVEMENT_PLAN.md)
3. **Provide feedback:**
   - Are these the right improvements?
   - Any concerns about the approaches?
   - Approve visualization approach (hybrid vs. alternatives)?
   - Any missing requirements?
4. **Prioritize** must-have vs. nice-to-have features
5. **Approve** to begin implementation

### For Implementation

Once approved:
1. Create feature branch
2. Set up development environment
3. Begin Phase 1 (Refactoring + Persistence)
4. Implement incrementally with testing
5. Gather feedback after each phase

---

## Questions for User

1. **Visualization:** Do you approve the hybrid approach (terminal + web dashboard)?
   - Alternative: Would you prefer static PNG export instead?
   
2. **Priorities:** Are all 4 improvements needed, or should we focus on a subset?
   
3. **Timeline:** Is the 3-4 week timeline acceptable?
   
4. **Optional Features:** Which optional features are worth including?
   
5. **Anything Unclear:** Any part of the plan that needs clarification?

---

## Contact

For questions or clarifications about this plan, please:
- Review the full plan: [IMPROVEMENT_PLAN.md](./IMPROVEMENT_PLAN.md)
- Check the research notes in the plan document
- Provide feedback on specific sections that need adjustment

---

**Ready to proceed pending your approval! 🚀**
