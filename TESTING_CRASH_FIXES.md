# Testing Crash Fixes

This document describes how to test the crash fixes implemented.

## Overview

The following issues were reported and fixed:
1. Program crashes when clicking "Save Config"
2. Program crashes when clicking column headers for sorting
3. Program crashes when clicking "Visualize" (with or without data)
4. Random text appears on console during crashes
5. No error logging for debugging

## Log File Location

All errors and panics are now logged to:
- Linux/Mac: `~/.config/manalyzer/manalyzer.log`
- Windows: `%APPDATA%\manalyzer\manalyzer.log`

The log file path is also displayed in the Event Log when the application starts.

## Testing Steps

### 1. Test "Save Config" Button

**Steps:**
1. Start the application: `./manalyzer`
2. Enter some player data (at least one name and SteamID)
3. Click "Save Config" button
4. Verify: App should NOT crash
5. Check Event Log for "Configuration saved successfully" message
6. Check log file for any errors

**Expected Result:**
- App continues running
- Config saved to `~/.config/manalyzer/config.json`
- No panic in log file

### 2. Test Column Header Sorting

**Prerequisites:** Must have analyzed some demo data first (or this feature won't be visible)

**Steps:**
1. Start the application
2. Run an analysis (if you have demo files)
3. Click on any stat column header (KAST%, ADR, K/D, etc.)
4. Verify: App should NOT crash
5. Verify: Table should resort by that column
6. Verify: Sort indicator (▲/▼) appears in header

**Expected Result:**
- App continues running
- Table data is sorted by clicked column
- Log file shows "Column X clicked for sorting"
- No panic in log file

### 3. Test "Visualize" Without Data

**Steps:**
1. Start fresh application: `./manalyzer`
2. Click "Visualize" button immediately (without running analysis)
3. Verify: App should NOT crash
4. Check Event Log for "Error: No data to visualize" message

**Expected Result:**
- App continues running
- Error message displayed in Event Log
- Log file shows "No data available for visualization"
- No panic in log file

### 4. Test "Visualize" With Data

**Prerequisites:** Must have demo data analyzed

**Steps:**
1. Start application
2. Run analysis on demo files
3. Click "Visualize" button
4. Verify: App should NOT crash
5. Browser should open (or error message if no browser available)

**Expected Result:**
- App continues running
- Browser opens to visualization dashboard OR
- Error message about browser opening failure
- Log file shows visualization server startup
- No panic in log file

## Checking Logs After Each Test

After each test, check the log file:

```bash
# View last 50 lines of log
tail -50 ~/.config/manalyzer/manalyzer.log

# Search for panics
grep -i panic ~/.config/manalyzer/manalyzer.log

# Search for errors
grep -i error ~/.config/manalyzer/manalyzer.log
```

## What Was Fixed

### 1. Added Comprehensive Logging (`src/logger.go`)
- All operations are now logged to file
- Panics are caught and logged with stack traces
- Log file created at startup

### 2. Added Panic Recovery
- `main.go`: Top-level panic recovery
- `gui_form.go`: Panic recovery in all button handlers
- `gui_table.go`: Panic recovery in column click handlers

### 3. Fixed Closure Bug
- Column header click handlers now properly capture the column index
- Changed from `columnIndex` to `sortCol` in closure to avoid variable capture issue

### 4. Added Better Error Handling
- Nil checks for statsTable before visualization
- Detailed logging at each step
- User-friendly error messages in Event Log

### 5. Improved Logging Coverage
- Startup/shutdown logged
- All button clicks logged
- Config operations logged
- Sort operations logged
- Visualization operations logged

## Common Issues

### "Permission Denied" on Log File
If you see permission errors, check:
```bash
ls -la ~/.config/manalyzer/
```
The directory and log file should be owned by your user.

### Log File Not Created
If the log file isn't created:
1. Check stdout for any errors during startup
2. Verify you have write permissions to ~/.config/
3. Check if the app actually started (might be another issue)

### Still Seeing Crashes
If you still see crashes after these fixes:
1. Check the log file immediately after crash
2. Look for the last PANIC entry
3. Copy the stack trace
4. Report the issue with the stack trace

## Success Criteria

All tests pass if:
- ✅ App never crashes unexpectedly
- ✅ Log file is created and contains detailed operation logs
- ✅ All panics are caught and logged
- ✅ User-friendly error messages shown in Event Log
- ✅ All button clicks work as expected
