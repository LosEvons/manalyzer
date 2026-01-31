package manalyzer

import (
"errors"
"fmt"
"time"

"github.com/akiver/cs-demo-analyzer/pkg/api"
)

var (
ErrNoValidSteamIDs = errors.New("no valid steamID64s found in any demos")
)

// PlayerStats represents statistics for a single player in a specific match on a specific side
type PlayerStats struct {
SteamID64  uint64
PlayerName string
Side       string // "T" or "CT"
// Statistics from STATISTICS.md
KAST            float32
ADR             float32
KillDeathRatio  float32
Kills           int
Deaths          int
FirstKills      int
FirstDeaths     int
TradeKills      int
TradeDeaths     int
HeadshotPercent int
AssistCount     int
}

// SideStats groups stats by side (T/CT) for a player in one match
type SideStats struct {
TStats  *PlayerStats // Stats when player was on T side
CTStats *PlayerStats // Stats when player was on CT side
}

// MatchStats groups all players' stats for a single match
type MatchStats struct {
MatchID     string                // Demo filename or checksum
MapName     string                // Map name
Date        time.Time             // Match date
PlayerStats map[uint64]*SideStats // keyed by steamID64
}

// MapStats groups matches by map name
type MapStats struct {
MapName string
Matches []*MatchStats
}

// AllPlayerStats is the top-level structure containing all matches grouped by map
type AllPlayerStats struct {
ByMap    map[string]*MapStats // keyed by map name
ByMatch  []*MatchStats        // all matches in order
NotFound []uint64             // steamID64s not found in any match
}

// FilterAndGroupPlayerStats processes matches and extracts player statistics
// grouped by match, map, and side (T/CT).
func FilterAndGroupPlayerStats(matches []*api.Match, steamIDs []uint64) (*AllPlayerStats, error) {
if len(steamIDs) == 0 {
return nil, errors.New("no steamID64s provided")
}

result := &AllPlayerStats{
ByMap:    make(map[string]*MapStats),
ByMatch:  make([]*MatchStats, 0),
NotFound: make([]uint64, 0),
}

// Track which steamIDs were found
foundIDs := make(map[uint64]bool)

// Process each match
for _, match := range matches {
matchStats := &MatchStats{
MatchID:     match.DemoFileName,
MapName:     match.MapName,
Date:        match.Date,
PlayerStats: make(map[uint64]*SideStats),
}

// Process each steamID
for _, steamID := range steamIDs {
player, exists := match.PlayersBySteamID[steamID]
if !exists {
continue
}

foundIDs[steamID] = true

// Extract stats for this player grouped by side
sideStats := extractPlayerStatsBySide(match, player)
matchStats.PlayerStats[steamID] = sideStats
}

// Only include matches where at least one target player was found
if len(matchStats.PlayerStats) > 0 {
result.ByMatch = append(result.ByMatch, matchStats)

// Group by map
if _, exists := result.ByMap[match.MapName]; !exists {
result.ByMap[match.MapName] = &MapStats{
MapName: match.MapName,
Matches: make([]*MatchStats, 0),
}
}
result.ByMap[match.MapName].Matches = append(result.ByMap[match.MapName].Matches, matchStats)
}
}

// Determine which steamIDs were not found
for _, steamID := range steamIDs {
if !foundIDs[steamID] {
result.NotFound = append(result.NotFound, steamID)
}
}

// Error if none of the steamIDs were found
if len(foundIDs) == 0 {
return result, ErrNoValidSteamIDs
}

return result, nil
}

// extractPlayerStatsBySide extracts player statistics separated by side (T/CT)
func extractPlayerStatsBySide(match *api.Match, player *api.Player) *SideStats {
result := &SideStats{}

// Determine which side the player started on
// Team A starts as CT, Team B starts as T
startsAsCT := player.Team.Name == match.TeamA.Name

// Calculate stats for first half (rounds 1-12) and second half (rounds 13+)
firstHalfStats := calculateStatsForRounds(match, player, 1, 12)
secondHalfStats := calculateStatsForRounds(match, player, 13, len(match.Rounds))

if startsAsCT {
result.CTStats = firstHalfStats
result.CTStats.Side = "CT"
result.TStats = secondHalfStats
result.TStats.Side = "T"
} else {
result.TStats = firstHalfStats
result.TStats.Side = "T"
result.CTStats = secondHalfStats
result.CTStats.Side = "CT"
}

return result
}

// calculateStatsForRounds calculates player statistics for a specific range of rounds
func calculateStatsForRounds(match *api.Match, player *api.Player, startRound, endRound int) *PlayerStats {
stats := &PlayerStats{
SteamID64:  player.SteamID64,
PlayerName: player.Name,
}

// Count statistics for the specified rounds
for _, event := range match.Kills {
roundNum := getRoundNumberForTick(match, event.Tick)
if roundNum < startRound || roundNum > endRound {
continue
}

// Check if player was the attacker
if event.KillerSteamID64 != 0 && event.KillerSteamID64 == player.SteamID64 {
stats.Kills++
if event.IsTradeKill {
stats.TradeKills++
}
}

// Check if player was the victim
if event.VictimSteamID64 != 0 && event.VictimSteamID64 == player.SteamID64 {
stats.Deaths++
if event.IsTradeDeath {
stats.TradeDeaths++
}
}
}

// Use player methods for overall stats, then scale by round count
totalRounds := float32(len(match.Rounds))
roundsInRange := float32(endRound - startRound + 1)
scaleFactor := roundsInRange / totalRounds

if totalRounds > 0 {
// Scale overall statistics to the round range
stats.KAST = player.KAST() * scaleFactor
stats.ADR = player.AverageDamagePerRound() * scaleFactor
stats.HeadshotPercent = player.HeadshotPercent()
stats.AssistCount = int(float32(player.AssistCount()) * scaleFactor)
stats.FirstKills = int(float32(player.FirstKillCount()) * scaleFactor)
stats.FirstDeaths = int(float32(player.FirstDeathCount()) * scaleFactor)
}

// Calculate K/D ratio
if stats.Deaths > 0 {
stats.KillDeathRatio = float32(stats.Kills) / float32(stats.Deaths)
} else if stats.Kills > 0 {
stats.KillDeathRatio = float32(stats.Kills)
}

return stats
}

// getRoundNumberForTick determines which round a tick belongs to
func getRoundNumberForTick(match *api.Match, tick int) int {
for _, round := range match.Rounds {
if tick >= round.StartTick && tick <= round.EndTick {
return round.Number
}
}
return 0
}

// FormatStatsForDisplay formats the AllPlayerStats into a readable string
func FormatStatsForDisplay(stats *AllPlayerStats) string {
if stats == nil {
return "No statistics available"
}

var result string

// Report not found steamIDs
if len(stats.NotFound) > 0 {
result += "Warning: The following steamID64s were not found:\n"
for _, id := range stats.NotFound {
result += fmt.Sprintf("  - %d\n", id)
}
result += "\n"
}

if len(stats.ByMatch) == 0 {
return result + "No matches found with the specified players."
}

result += fmt.Sprintf("Found %d match(es) with specified players.\n\n", len(stats.ByMatch))

// Display stats grouped by map
for mapName, mapStats := range stats.ByMap {
result += fmt.Sprintf("=== Map: %s (%d matches) ===\n\n", mapName, len(mapStats.Matches))

for _, match := range mapStats.Matches {
result += fmt.Sprintf("Match: %s (Date: %s)\n", match.MatchID, match.Date.Format("2006-01-02 15:04"))

for steamID, sideStats := range match.PlayerStats {
result += fmt.Sprintf("\n  Player: %s (SteamID: %d)\n",
getSteamIDPlayerName(sideStats), steamID)

// Display CT stats
if sideStats.CTStats != nil {
result += "    [CT Side]\n"
result += formatPlayerStats(sideStats.CTStats)
}

// Display T stats
if sideStats.TStats != nil {
result += "    [T Side]\n"
result += formatPlayerStats(sideStats.TStats)
}
}
result += "\n"
}
}

return result
}

// getSteamIDPlayerName extracts player name from SideStats
func getSteamIDPlayerName(sideStats *SideStats) string {
if sideStats.CTStats != nil {
return sideStats.CTStats.PlayerName
}
if sideStats.TStats != nil {
return sideStats.TStats.PlayerName
}
return "Unknown"
}

// formatPlayerStats formats a single PlayerStats struct
func formatPlayerStats(stats *PlayerStats) string {
if stats == nil {
return "      No data\n"
}

return fmt.Sprintf(
"      K: %d | D: %d | K/D: %.2f | ADR: %.1f | KAST: %.1f%% | HS%%: %d\n"+
"      FK: %d | FD: %d | TK: %d | TD: %d | Assists: %d\n",
stats.Kills, stats.Deaths, stats.KillDeathRatio, stats.ADR, stats.KAST*100, stats.HeadshotPercent,
stats.FirstKills, stats.FirstDeaths, stats.TradeKills, stats.TradeDeaths, stats.AssistCount,
)
}
