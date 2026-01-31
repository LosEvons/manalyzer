# Manalyzer - Complete Implementation Guide

**Status:** APPROVED FOR IMPLEMENTATION  
**Date:** 2026-01-31  
**Approach:** Hybrid Visualization (Terminal + Web Dashboard)

---

## Quick Reference

This guide provides step-by-step instructions for implementing all 4 approved improvements. See PLAN_CORRECTIONS.md for detailed code examples and PLAN_AUDIT.md for identified issues that were fixed.

### Approved Configuration

✅ **Visualization:** Hybrid (Terminal + Web Dashboard with go-echarts)  
✅ **Optional Features:** Save Config button, sort indicators, browser fallbacks, dynamic ports  
⚠️ **Deferred:** Stat range filters, alternative grouping, chart export, large function refactoring

### Implementation Order

1. Phase 1: Refactoring (Week 1, Day 1-2)
2. Phase 2: Persistent Storage (Week 1, Day 3-4)  
3. Phase 3: Filtering & Sorting (Week 2)
4. Phase 4: Visualization (Week 3-4)

### Total Effort: 22-30 hours over 3-4 weeks

---

## Phase 1: Refactoring

**Goal:** Split gui.go into focused modules.

**New Files to Create:**
- `src/gui_eventlog.go` - EventLog component
- `src/gui_form.go` - Form and handlers
- `src/gui_table.go` - StatisticsTable with sorting
- `src/config.go` - Configuration structures

**See PLAN_CORRECTIONS.md Section 1 for complete code.**

---

## Phase 2: Persistent Storage

**Goal:** Save/load player configuration.

**Key Changes:**
- Add `config *Config` field to UI struct
- Load config in New()
- Save config in onAnalyzeClicked
- Add "Save Config" button

**See PLAN_CORRECTIONS.md Section 2-4 for implementation details.**

---

## Phase 3: Filtering & Sorting

**Goal:** Interactive table controls.

**Key Features:**
- Clickable column headers for sorting
- Sort indicators (▲/▼)
- Map filter dropdown
- Side filter dropdown (T/CT/All)

**See PLAN_CORRECTIONS.md Section 3 and gui_table.go code.**

---

## Phase 4: Visualization

**Goal:** Web dashboard with go-echarts.

**Key Components:**
- StartVisualizationServer with port handling
- openBrowser with fallbacks
- Chart handlers (player comparison, T vs CT, map breakdown)
- "Visualize" button in GUI

**See PLAN_CORRECTIONS.md Section 5-7 for complete code.**

---

## Critical Fixes Applied

The following issues from PLAN_AUDIT.md were corrected:

### Fixed in PLAN_CORRECTIONS.md:
1. ✅ UI struct now includes `config *Config` field
2. ✅ StatisticsTable includes `app`, `sortColumn`, `sortDesc` fields
3. ✅ Function signatures properly migrated (createPlayerInputFormWithConfig)
4. ✅ toggleSort() fully implemented
5. ✅ sortData() fully implemented  
6. ✅ Header click handler closure bug fixed
7. ✅ Browser auto-open with multiple fallbacks
8. ✅ Port conflict handling (8080-8090)
9. ✅ buildConfigFromForm() added

---

## Implementation Checklist

### Pre-Implementation
- [ ] Create feature branch: `git checkout -b feature/improvements`
- [ ] Review PLAN_CORRECTIONS.md for all code examples
- [ ] Set up test environment with sample .dem files

### Phase 1: Refactoring (2 days)
- [ ] Create gui_eventlog.go from PLAN_CORRECTIONS.md Section 1.1
- [ ] Compile and test
- [ ] Create gui_table.go with sorting from Section 1.2
- [ ] Compile and test
- [ ] Create gui_form.go from Section 1.3
- [ ] Compile and test
- [ ] Update gui.go to core functions only from Section 1.4
- [ ] Create config.go from Section 1.5
- [ ] Full regression test
- [ ] Commit: "refactor: split gui.go into focused modules"

### Phase 2: Persistence (2 days)
- [ ] Verify config structures from Phase 1
- [ ] Test config save/load manually
- [ ] Update New() to load config (already in Phase 1)
- [ ] Test first run (no config)
- [ ] Test save and reload
- [ ] Test manual JSON edits
- [ ] Commit: "feat: add persistent configuration"

### Phase 3: Filtering & Sorting (1 week)
- [ ] Verify sorting works (added in Phase 1)
- [ ] Test all column sorts
- [ ] Add filter dropdowns (code in PLAN_CORRECTIONS.md)
- [ ] Test map filtering
- [ ] Test side filtering
- [ ] Test combined filters with sorting
- [ ] Commit: "feat: add interactive filtering and sorting"

### Phase 4: Visualization (2 weeks)
- [ ] Add go-echarts dependency: `go get github.com/go-echarts/go-echarts/v2`
- [ ] Create visualize.go from Section 5
- [ ] Implement all chart handlers from Section 6
- [ ] Add "Visualize" button from Section 7
- [ ] Test on Linux
- [ ] Test on macOS  
- [ ] Test on Windows
- [ ] Test edge cases (no browser, port conflicts)
- [ ] Commit: "feat: add web visualization dashboard"

### Final Steps
- [ ] Update README.md with new features
- [ ] Run full regression test
- [ ] Test with large dataset (100+ demos)
- [ ] Create PR
- [ ] Tag release v1.1.0

---

## Verification Steps

See PLAN_CORRECTIONS.md Section 10 for detailed verification procedures for each phase.

**Quick Checks:**
- Phase 1: App compiles and works identically
- Phase 2: Config persists between sessions
- Phase 3: Can sort and filter table data
- Phase 4: "Visualize" button opens browser with charts

---

## Common Pitfalls

See PLAN_CORRECTIONS.md Section 9 for detailed pitfall descriptions.

**Top 3 to Avoid:**
1. ❌ Don't capture loop variable in closure - use `columnIndex := col`
2. ❌ Don't forget QueueUpdateDraw for UI updates
3. ❌ Don't assume browser is always available - provide fallback URL

---

## Success Criteria

✅ All 4 improvements implemented
✅ Included optional features work
✅ No existing functionality broken
✅ Config persists across sessions
✅ Sorting and filtering work together
✅ Visualization opens in browser
✅ Cross-platform compatible

---

## Key Documents Reference

- **IMPROVEMENT_PLAN.md** - Original research and recommendations
- **IMPROVEMENT_SUMMARY.md** - Executive summary
- **PLAN_AUDIT.md** - Issues found in original plan
- **PLAN_CORRECTIONS.md** - ⭐ Complete corrected code examples
- **This file** - Implementation roadmap and checklist

---

**Status: READY FOR IMPLEMENTATION** ✅

All planning complete. All issues identified and fixed. Complete code examples provided in PLAN_CORRECTIONS.md.

An AI agent can now implement all 4 improvements by following the step-by-step instructions in PLAN_CORRECTIONS.md sections 1-7, using this file's checklist to track progress.
