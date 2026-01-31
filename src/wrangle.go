package manalyzer

import (
	"fmt"
	"strconv"

	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

// PlayerStats holds all statistics for ONE player across all matches
type PlayerStats struct {
	SteamID64 string // Primary key
	PlayerName string // For display only

	// Per-map statistics with T/CT breakdown
	MapStats map[string]*MapStatistics

	// Overall statistics (aggregated across all maps and sides)
	OverallStats *OverallStatistics
}

// MapStatistics holds stats for ONE player on ONE map
type MapStatistics struct {
	MapName       string
	MatchesPlayed int

	// Statistics separated by side (T = Terrorist, CT = Counter-Terrorist)
	SideStats map[string]*SideStatistics // Keys: "T" and "CT"
}

// SideStatistics holds stats for ONE player on ONE map on ONE side
type SideStatistics struct {
	Side string // "T" or "CT"

	// Core statistics from STATISTICS.md
	KAST        float64 // Percentage (0-100)
	ADR         float64 // Average Damage per Round
	KD          float64 // Kill/Death ratio
	Kills       int
	Deaths      int
	FirstKills  int // First kill of the round
	FirstDeaths int // First death of the round
	TradeKills  int // Killed enemy shortly after teammate death
	TradeDeaths int // Killed shortly after getting a kill

	// Additional useful stats
	Assists      int
	Headshots    int
	RoundsPlayed int
}

// OverallStatistics holds aggregated stats across ALL maps and sides
type OverallStatistics struct {
	KAST          float64 // Weighted average by rounds
	ADR           float64 // Weighted average by rounds
	KD            float64 // Total kills / total deaths
	Kills         int     // Sum across all maps/sides
	Deaths        int     // Sum
	FirstKills    int     // Sum
	FirstDeaths   int     // Sum
	TradeKills    int     // Sum
	TradeDeaths   int     // Sum
	Assists       int     // Sum
	Headshots     int     // Sum
	RoundsPlayed  int     // Sum
	MatchesPlayed int     // Count of unique matches
}

// WrangleResult is the final output of ProcessMatches()
type WrangleResult struct {
	PlayerStats  []*PlayerStats
	MapList      []string // Unique map names (for UI filtering)
	TotalMatches int
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// determinePlayerSideInRound returns which side (T or CT) a player was on in a specific round
func determinePlayerSideInRound(match *api.Match, player *api.Player, round *api.Round) common.Team {
	// Player belongs to either TeamA or TeamB
	// Each round records which side TeamA and TeamB were on
	if player.Team == match.TeamA {
		return round.TeamASide
	}
	return round.TeamBSide
}

// sideToString converts common.Team to "T" or "CT" string for map keys
// Returns empty string for unassigned/spectator teams
func sideToString(side common.Team) string {
	if side == common.TeamTerrorists {
		return "T"
	} else if side == common.TeamCounterTerrorists {
		return "CT"
	}
	// Return empty string for TeamUnassigned, TeamSpectators, etc.
	return ""
}

// ============================================================================
// CORE FUNCTIONS
// ============================================================================

// ProcessMatches is the main entry point for data processing
func ProcessMatches(matches []*api.Match, steamIDs []string) (*WrangleResult, error) {
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches to process")
	}

	// Convert string SteamIDs to uint64
	steamID64s := make([]uint64, 0, len(steamIDs))
	for _, steamIDStr := range steamIDs {
		if steamIDStr == "" {
			continue
		}
		steamID64, err := strconv.ParseUint(steamIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid SteamID64 %s: %w", steamIDStr, err)
		}
		steamID64s = append(steamID64s, steamID64)
	}

	if len(steamID64s) == 0 {
		return nil, fmt.Errorf("no valid SteamIDs provided")
	}

	// Initialize PlayerStats for each SteamID
	playerStatsMap := make(map[uint64]*PlayerStats)
	for _, steamID64 := range steamID64s {
		playerStatsMap[steamID64] = &PlayerStats{
			SteamID64: strconv.FormatUint(steamID64, 10),
			MapStats:  make(map[string]*MapStatistics),
		}
	}

	// Track unique maps
	mapsEncountered := make(map[string]bool)

	// Process each match
	for _, match := range matches {
		mapName := match.MapName
		mapsEncountered[mapName] = true

		for steamID64, playerStats := range playerStatsMap {
			// Check if player exists in this match
			player, exists := match.PlayersBySteamID[steamID64]
			if !exists {
				continue // Player wasn't in this match
			}

			// Set player name (for display) from first match
			if playerStats.PlayerName == "" {
				playerStats.PlayerName = player.Name
			}

			// Initialize map stats if needed
			if playerStats.MapStats[mapName] == nil {
				playerStats.MapStats[mapName] = &MapStatistics{
					MapName:       mapName,
					MatchesPlayed: 0,
					SideStats:     make(map[string]*SideStatistics),
				}
			}

			mapStats := playerStats.MapStats[mapName]
			mapStats.MatchesPlayed++

			// Extract stats by side
			sideStatsFromMatch := extractPlayerStatsBySide(match, player)

			// Aggregate into existing stats
			for sideKey, newStats := range sideStatsFromMatch {
				if mapStats.SideStats[sideKey] == nil {
					mapStats.SideStats[sideKey] = &SideStatistics{Side: sideKey}
				}

				existing := mapStats.SideStats[sideKey]

				// Sum counts
				existing.Kills += newStats.Kills
				existing.Deaths += newStats.Deaths
				existing.Assists += newStats.Assists
				existing.FirstKills += newStats.FirstKills
				existing.FirstDeaths += newStats.FirstDeaths
				existing.TradeKills += newStats.TradeKills
				existing.TradeDeaths += newStats.TradeDeaths
				existing.Headshots += newStats.Headshots

				// Weighted average for ADR
				oldRounds := existing.RoundsPlayed
				newRounds := newStats.RoundsPlayed
				existing.RoundsPlayed += newRounds

				if existing.RoundsPlayed > 0 {
					oldDamage := existing.ADR * float64(oldRounds)
					newDamage := newStats.ADR * float64(newRounds)
					existing.ADR = (oldDamage + newDamage) / float64(existing.RoundsPlayed)
				}

				// Weighted average for KAST
				if existing.RoundsPlayed > 0 {
					oldKAST := (existing.KAST / 100.0) * float64(oldRounds)
					newKAST := (newStats.KAST / 100.0) * float64(newRounds)
					existing.KAST = ((oldKAST + newKAST) / float64(existing.RoundsPlayed)) * 100.0
				}

				// Recalculate K/D
				if existing.Deaths > 0 {
					existing.KD = float64(existing.Kills) / float64(existing.Deaths)
				} else if existing.Kills > 0 {
					existing.KD = float64(existing.Kills)
				}
			}
		}
	}

	// Calculate overall stats for each player
	for _, playerStats := range playerStatsMap {
		playerStats.OverallStats = calculateOverallStats(playerStats.MapStats)
	}

	// Convert to slice and return
	playerStatsList := make([]*PlayerStats, 0, len(playerStatsMap))
	for _, stats := range playerStatsMap {
		playerStatsList = append(playerStatsList, stats)
	}

	mapList := make([]string, 0, len(mapsEncountered))
	for mapName := range mapsEncountered {
		mapList = append(mapList, mapName)
	}

	return &WrangleResult{
		PlayerStats:  playerStatsList,
		MapList:      mapList,
		TotalMatches: len(matches),
	}, nil
}

// extractPlayerStatsBySide analyzes a match and extracts side-specific statistics for a player
func extractPlayerStatsBySide(match *api.Match, player *api.Player) map[string]*SideStatistics {
	// Initialize stats for both sides
	sideStats := make(map[string]*SideStatistics)
	sideStats["T"] = &SideStatistics{Side: "T"}
	sideStats["CT"] = &SideStatistics{Side: "CT"}

	// Count rounds per side
	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue // Skip unassigned teams
		}
		sideStats[sideKey].RoundsPlayed++
	}

	// Process kills and deaths
	for _, kill := range match.Kills {
		// Find which round this kill happened in
		var round *api.Round
		for _, r := range match.Rounds {
			if r.Number == kill.RoundNumber {
				round = r
				break
			}
		}
		if round == nil {
			continue
		}

		// Determine player's side in this round
		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue // Skip unassigned teams
		}
		stats := sideStats[sideKey]

		// Count kills (if player is killer)
		if kill.KillerSteamID64 == player.SteamID64 && !kill.IsKillerControllingBot {
			if !kill.IsSuicide() && !kill.IsTeamKill() {
				stats.Kills++
				if kill.IsHeadshot {
					stats.Headshots++
				}
				if kill.IsTradeKill {
					stats.TradeKills++
				}
			}
		}

		// Count deaths (if player is victim)
		if kill.VictimSteamID64 == player.SteamID64 && !kill.IsVictimControllingBot {
			if !kill.IsSuicide() {
				stats.Deaths++
				if kill.IsTradeDeath {
					stats.TradeDeaths++
				}
			}
		}

		// Count assists
		if kill.AssisterSteamID64 == player.SteamID64 && !kill.IsAssisterControllingBot {
			if kill.AssisterSide != kill.VictimSide {
				stats.Assists++
			}
		}
	}

	// Calculate first kills/deaths per side
	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue // Skip unassigned teams
		}
		stats := sideStats[sideKey]

		// Get all kills in this round
		var killsInRound []*api.Kill
		for _, kill := range match.Kills {
			if kill.RoundNumber == round.Number {
				killsInRound = append(killsInRound, kill)
			}
		}

		// Find first kill
		for _, kill := range killsInRound {
			if kill.IsKillerControllingBot || kill.IsSuicide() || kill.IsTeamKill() {
				continue
			}
			// This is the first valid kill in the round
			if kill.KillerSteamID64 == player.SteamID64 {
				stats.FirstKills++
			}
			break // Stop after finding first valid kill
		}

		// Find first death
		for _, kill := range killsInRound {
			if kill.IsVictimControllingBot || kill.IsSuicide() || kill.IsTeamKill() {
				continue
			}
			// This is the first valid death in the round
			if kill.VictimSteamID64 == player.SteamID64 {
				stats.FirstDeaths++
			}
			break // Stop after finding first valid death
		}
	}

	// Calculate damage and ADR
	totalDamagePerSide := make(map[string]int)
	for _, damage := range match.Damages {
		if damage.AttackerSteamID64 != player.SteamID64 {
			continue
		}

		// Find which round this damage occurred in
		for _, round := range match.Rounds {
			if damage.Tick >= round.StartTick && damage.Tick <= round.EndTick {
				playerSide := determinePlayerSideInRound(match, player, round)
				sideKey := sideToString(playerSide)
				if sideKey != "" {
					totalDamagePerSide[sideKey] += damage.HealthDamage
				}
				break
			}
		}
	}

	// Calculate ADR (Average Damage per Round)
	for sideKey, totalDamage := range totalDamagePerSide {
		if sideKey == "" {
			continue // Skip unassigned
		}
		if stats, ok := sideStats[sideKey]; ok && stats != nil && stats.RoundsPlayed > 0 {
			stats.ADR = float64(totalDamage) / float64(stats.RoundsPlayed)
		}
	}

	// Calculate K/D for each side
	for _, stats := range sideStats {
		if stats.Deaths > 0 {
			stats.KD = float64(stats.Kills) / float64(stats.Deaths)
		} else if stats.Kills > 0 {
			stats.KD = float64(stats.Kills)
		}
	}

	// Calculate KAST for each side
	sideStats["T"].KAST = calculateKASTForSide(match, player, common.TeamTerrorists)
	sideStats["CT"].KAST = calculateKASTForSide(match, player, common.TeamCounterTerrorists)

	return sideStats
}

// calculateKASTForSide calculates KAST percentage for a specific side
// KAST = (Kill or Assist or Survived or Traded) / Total Rounds
// Returns percentage (0-100)
func calculateKASTForSide(match *api.Match, player *api.Player, side common.Team) float64 {
	kastPerRound := make(map[int]bool)
	roundsOnThisSide := 0

	// Check each round
	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		if playerSide != side {
			continue // Player was on opposite side this round
		}

		roundsOnThisSide++
		kastPerRound[round.Number] = false
		playerSurvived := true

		// Check all kills in this round
		for _, kill := range match.Kills {
			if round.Number != kill.RoundNumber {
				continue
			}

			// Skip team kills
			isTeamKill := kill.KillerSide == kill.VictimSide
			if isTeamKill {
				continue
			}

			// Check for Assist
			if kill.AssisterSteamID64 == player.SteamID64 {
				kastPerRound[round.Number] = true
			}

			// Check for Kill
			if kill.KillerSteamID64 == player.SteamID64 && kill.VictimSteamID64 != player.SteamID64 {
				kastPerRound[round.Number] = true
			}

			// Check for Death
			if kill.VictimSteamID64 == player.SteamID64 {
				playerSurvived = false
				// Check if Traded
				if kill.IsTradeDeath {
					kastPerRound[round.Number] = true
				}
			}
		}

		// Check for Survived
		if playerSurvived {
			kastPerRound[round.Number] = true
		}
	}

	// Calculate percentage
	kastEventCount := 0
	for _, hasKASTEvent := range kastPerRound {
		if hasKASTEvent {
			kastEventCount++
		}
	}

	if roundsOnThisSide > 0 {
		return (float64(kastEventCount) / float64(roundsOnThisSide)) * 100.0
	}

	return 0.0
}

// calculateOverallStats aggregates statistics across all maps and sides
func calculateOverallStats(mapStats map[string]*MapStatistics) *OverallStatistics {
	overall := &OverallStatistics{}

	// Sum all counts
	for _, mapStat := range mapStats {
		overall.MatchesPlayed += mapStat.MatchesPlayed

		for _, sideStat := range mapStat.SideStats {
			overall.Kills += sideStat.Kills
			overall.Deaths += sideStat.Deaths
			overall.Assists += sideStat.Assists
			overall.FirstKills += sideStat.FirstKills
			overall.FirstDeaths += sideStat.FirstDeaths
			overall.TradeKills += sideStat.TradeKills
			overall.TradeDeaths += sideStat.TradeDeaths
			overall.Headshots += sideStat.Headshots
			overall.RoundsPlayed += sideStat.RoundsPlayed
		}
	}

	// Calculate K/D
	if overall.Deaths > 0 {
		overall.KD = float64(overall.Kills) / float64(overall.Deaths)
	} else if overall.Kills > 0 {
		overall.KD = float64(overall.Kills)
	}

	// Weighted average for ADR
	totalDamage := 0.0
	for _, mapStat := range mapStats {
		for _, sideStat := range mapStat.SideStats {
			totalDamage += sideStat.ADR * float64(sideStat.RoundsPlayed)
		}
	}
	if overall.RoundsPlayed > 0 {
		overall.ADR = totalDamage / float64(overall.RoundsPlayed)
	}

	// Weighted average for KAST
	kastRoundsTotal := 0.0
	for _, mapStat := range mapStats {
		for _, sideStat := range mapStat.SideStats {
			kastRoundsTotal += (sideStat.KAST / 100.0) * float64(sideStat.RoundsPlayed)
		}
	}
	if overall.RoundsPlayed > 0 {
		overall.KAST = (kastRoundsTotal / float64(overall.RoundsPlayed)) * 100.0
	}

	return overall
}
