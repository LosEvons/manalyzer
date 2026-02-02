# Manalyzer

CS:GO demo analyzer with a terminal UI for tracking player statistics with side-specific breakdowns.

## Features

- **Player Statistics Tracking**: Analyze up to 5 players simultaneously by their SteamID64
- **Side-Specific Stats**: View statistics broken down by Terrorist (T) and Counter-Terrorist (CT) sides
- **Map-Based Analysis**: See performance across different maps
- **Comprehensive Metrics**: KAST, ADR, K/D, Kills, Deaths, First Kills/Deaths, Trade Kills/Deaths
- **Recursive Demo Scanning**: Automatically finds all .dem files in a directory tree
- **Interactive TUI**: Easy-to-use terminal interface with real-time event logging
- **Persistent Configuration**: Auto-saves player configurations between sessions
- **Interactive Sorting**: Click column headers to sort statistics
- **Web Visualization**: Browser-based interactive charts with dark theme
- **Robust Error Handling**: Graceful error recovery with clear user feedback

## Requirements

- Go 1.23 or higher
- CS:GO demo files (.dem format)

## Installation

```bash
# Clone the repository
git clone https://github.com/LosEvons/manalyzer.git
cd manalyzer

# Build the application
go build

# Run the application
./manalyzer
```

## Usage

1. **Launch the Application**: Run `./manalyzer` to start the terminal UI

2. **Configure Players**:
   - Enter player names (optional, for display purposes)
   - Enter SteamID64 values (17-digit numbers) for each player you want to track
   - You can track 1-5 players at a time

3. **Set Demo Path**:
   - Enter the path to a directory containing CS:GO demo files
   - The application will recursively search for all `.dem` files

4. **Analyze**:
   - Click the "Analyze" button to start processing demos
   - Watch the Event Log for progress updates
   - View results in the Statistics Table below
   - Player configurations are automatically saved

5. **Sort Statistics**:
   - Click any stat column header to sort (KAST%, ADR, K/D, etc.)
   - Click again to reverse sort direction
   - Visual indicators (▲/▼) show current sort

6. **Visualize Data**:
   - Click the "Visualize" button after analysis
   - Opens browser with interactive charts
   - View player comparisons, T vs CT performance, and map breakdowns

7. **Save Configuration**:
   - Use the "Save Config" button to manually save player settings
   - Configuration auto-saves after each analysis

8. **Clear Form**:
   - Use the "Clear" button to reset all input fields

## Statistics Explained

- **KAST**: Percentage of rounds where you got a Kill, Assist, Survived, or were Traded (0-100%)
- **ADR**: Average Damage per Round
- **K/D**: Kill/Death ratio
- **FK/FD**: First Kills / First Deaths (first kill/death of the round)
- **TK/TD**: Trade Kills / Trade Deaths (TK: killing an enemy shortly after a teammate's death; TD: being killed shortly after a teammate's death)

## Interface Layout

```
┌─────────────────────────────┬─────────────────────────────┐
│  Player Configuration       │   Event Log (5 rows)        │
│                             ├─────────────────────────────┤
│  • Player 1-5 inputs        │                             │
│  • SteamID64 fields         │   Statistics Table          │
│  • Demo path                │   (sortable by clicking     │
│  • [Analyze] [Clear]        │    column headers)          │
│  • [Save Config]            │                             │
│  • [Visualize]              │   ▲/▼ Sort indicators       │
└─────────────────────────────┴─────────────────────────────┘
```

## Controls

- **ESC** or **Ctrl+C**: Exit the application
- **Tab**: Navigate between form fields
- **Enter**: Activate buttons or submit fields
- **Mouse**: Click column headers to sort, click buttons

## Configuration

Player configurations are automatically saved to:
- Linux/macOS: `~/.config/manalyzer/config.json`
- Fallback: `~/.manalyzer/config.json`

The configuration persists between sessions and includes:
- Player names and SteamID64 values
- Last used demo path
- Preferences (auto-save enabled by default)

## Technical Details

### Data Structures

- **Side-Specific Statistics**: All statistics are calculated separately for T and CT sides
- **Map-Based Grouping**: Statistics are organized by map for detailed analysis
- **Overall Aggregation**: Weighted averages computed across all matches and sides

### Demo Processing

- Uses `cs-demo-analyzer` library for parsing CS:GO demos
- Supports Valve demo format
- Handles corrupted demos gracefully (continues processing others)
- Logs detailed progress and errors
- All errors displayed in Event Log (no crashes)

### Error Handling

- **Panic Recovery**: All goroutines protected with panic recovery
- **Graceful Degradation**: Continues operation when individual demos fail
- **Clear Feedback**: All errors displayed in Event Log with timestamps
- **No Crashes**: Application cannot crash uncontrollably

## Visualization

After running analysis, click "Visualize" to open a web dashboard with:

1. **Player Comparison** - Bar chart comparing overall performance (KAST%, ADR, K/D)
2. **T vs CT Performance** - Side-specific KAST% comparison
3. **Map Breakdown** - Performance by map for each player

The visualization server runs on `localhost:8080-8090` (auto-selects available port) and opens automatically in your default browser.

## Development

Built with:
- [cs-demo-analyzer](https://github.com/akiver/cs-demo-analyzer) - Demo parsing
- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [tcell](https://github.com/gdamore/tcell) - Terminal handling
- [go-echarts](https://github.com/go-echarts/go-echarts) - Data visualization

## License

See LICENSE file for details.
