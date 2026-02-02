# Manalyzer

CS:GO demo analyzer with a terminal UI for tracking player statistics with side-specific breakdowns.

## Features

- **Player Statistics Tracking**: Analyze up to 5 players simultaneously by their SteamID64
- **Side-Specific Stats**: View statistics broken down by Terrorist (T) and Counter-Terrorist (CT) sides
- **Map-Based Analysis**: See performance across different maps
- **Comprehensive Metrics**: KAST, ADR, K/D, Kills, Deaths, First Kills/Deaths, Trade Kills/Deaths
- **Recursive Demo Scanning**: Automatically finds all .dem files in a directory tree
- **Interactive TUI**: Easy-to-use terminal interface with real-time event logging
- **Persistent Configuration**: Auto-save/load player configurations between sessions
- **Interactive Sorting**: Click column headers to sort statistics by any metric
- **Web Visualization**: View interactive charts and graphs in your browser

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
   - Click column headers to sort by any metric

5. **Save Configuration**:
   - Click "Save Config" to manually save player settings
   - Configuration auto-saves after each analysis
   - Config stored in `~/.config/manalyzer/config.json`

6. **Visualize Results**:
   - Click "Visualize" to open web dashboard
   - View interactive charts in your browser
   - Three chart types: Player Comparison, T vs CT Performance, Map Breakdown

7. **Clear Form**:
   - Use the "Clear" button to reset all input fields

## Statistics Explained

- **KAST**: Percentage of rounds where you got a Kill, Assist, Survived, or were Traded (0-100%)
- **ADR**: Average Damage per Round
- **K/D**: Kill/Death ratio
- **FK/FD**: First Kills / First Deaths (first kill/death of the round)
- **TK/TD**: Trade Kills / Trade Deaths (TK: killing an enemy shortly after a teammate's death; TD: being killed shortly after a teammate's death)

## Interface Layout

```
┌─────────────────────────┬───────────────────────────────┐
│  Player Configuration   │      Event Log (5 rows)       │
│                         ├───────────────────────────────┤
│  • Player 1-5 inputs    │                               │
│  • SteamID64 fields     │    Statistics Table           │
│  • Demo path            │    (Click headers to sort)    │
│  • [Analyze] [Clear]    │                               │
│  • [Save Config]        │                               │
│  • [Visualize]          │                               │
└─────────────────────────┴───────────────────────────────┘
```

## Controls

- **ESC** or **Ctrl+C**: Exit the application
- **Tab**: Navigate between form fields
- **Enter**: Activate buttons or submit fields

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

## Development

Built with:
- [cs-demo-analyzer](https://github.com/akiver/cs-demo-analyzer) - Demo parsing
- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [tcell](https://github.com/gdamore/tcell) - Terminal handling
- [go-echarts](https://github.com/go-echarts/go-echarts) - Interactive charts

## License

See LICENSE file for details.
