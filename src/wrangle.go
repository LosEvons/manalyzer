package manalyzer

import (
	"fmt"
	"strconv"

	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

// PlayerStats holds statistics for a player across all matches.
type PlayerStats struct {
	SteamID64    string
	PlayerName   string
	MapStats     map[string]*MapStatistics
	OverallStats *OverallStatistics
}

// MapStatistics holds per-map statistics for a player.
type MapStatistics struct {
	MapName       string
	MatchesPlayed int
	SideStats     map[string]*SideStatistics // Keys: "T" and "CT"
}

// SideStatistics holds statistics for one side (T or CT) on a map.
type SideStatistics struct {
	Side         string
	KAST         float64 // Percentage (0-100)
	ADR          float64
	KD           float64
	Kills        int
	Deaths       int
	FirstKills   int
	FirstDeaths  int
	TradeKills   int
	TradeDeaths  int
	Assists      int
	Headshots    int
	RoundsPlayed int
}

// OverallStatistics holds aggregated stats across all maps and sides.
type OverallStatistics struct {
	KAST          float64
	ADR           float64
	KD            float64
	Kills         int
	Deaths        int
	FirstKills    int
	FirstDeaths   int
	TradeKills    int
	TradeDeaths   int
	Assists       int
	Headshots     int
	RoundsPlayed  int
	MatchesPlayed int
}

// WrangleResult is the output of ProcessMatches.
type WrangleResult struct {
	PlayerStats  []*PlayerStats
	MapList      []string
	TotalMatches int
}

// determinePlayerSideInRound returns which side (T or CT) a player was on.
func determinePlayerSideInRound(match *api.Match, player *api.Player, round *api.Round) common.Team {
	if player.Team == match.TeamA {
		return round.TeamASide
	}
	return round.TeamBSide
}

// sideToString converts common.Team to "T" or "CT" string.
// Returns empty string for unassigned/spectator teams.
func sideToString(side common.Team) string {
	if side == common.TeamTerrorists {
		return "T"
	} else if side == common.TeamCounterTerrorists {
		return "CT"
	}
	return ""
}

// ProcessMatches processes demo matches and extracts player statistics.
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

	playerStatsMap := make(map[uint64]*PlayerStats)
	for _, steamID64 := range steamID64s {
		playerStatsMap[steamID64] = &PlayerStats{
			SteamID64: strconv.FormatUint(steamID64, 10),
			MapStats:  make(map[string]*MapStatistics),
		}
	}

	mapsEncountered := make(map[string]bool)

	for _, match := range matches {
		mapName := match.MapName
		mapsEncountered[mapName] = true

		for steamID64, playerStats := range playerStatsMap {
			player, exists := match.PlayersBySteamID[steamID64]
			if !exists {
				continue
			}

			if playerStats.PlayerName == "" {
				playerStats.PlayerName = player.Name
			}

			if playerStats.MapStats[mapName] == nil {
				playerStats.MapStats[mapName] = &MapStatistics{
					MapName:       mapName,
					MatchesPlayed: 0,
					SideStats:     make(map[string]*SideStatistics),
				}
			}

			mapStats := playerStats.MapStats[mapName]
			mapStats.MatchesPlayed++

			sideStatsFromMatch := extractPlayerStatsBySide(match, player)

			for sideKey, newStats := range sideStatsFromMatch {
				if mapStats.SideStats[sideKey] == nil {
					mapStats.SideStats[sideKey] = &SideStatistics{Side: sideKey}
				}

				existing := mapStats.SideStats[sideKey]

				existing.Kills += newStats.Kills
				existing.Deaths += newStats.Deaths
				existing.Assists += newStats.Assists
				existing.FirstKills += newStats.FirstKills
				existing.FirstDeaths += newStats.FirstDeaths
				existing.TradeKills += newStats.TradeKills
				existing.TradeDeaths += newStats.TradeDeaths
				existing.Headshots += newStats.Headshots

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

	for _, playerStats := range playerStatsMap {
		playerStats.OverallStats = calculateOverallStats(playerStats.MapStats)
	}

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

// extractPlayerStatsBySide extracts side-specific statistics for a player from a match.
func extractPlayerStatsBySide(match *api.Match, player *api.Player) map[string]*SideStatistics {
	sideStats := make(map[string]*SideStatistics)
	sideStats["T"] = &SideStatistics{Side: "T"}
	sideStats["CT"] = &SideStatistics{Side: "CT"}

	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue
		}
		sideStats[sideKey].RoundsPlayed++
	}

	for _, kill := range match.Kills {
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

		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue
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

		if kill.VictimSteamID64 == player.SteamID64 && !kill.IsVictimControllingBot {
			if !kill.IsSuicide() {
				stats.Deaths++
				if kill.IsTradeDeath {
					stats.TradeDeaths++
				}
			}
		}

		if kill.AssisterSteamID64 == player.SteamID64 && !kill.IsAssisterControllingBot {
			if kill.AssisterSide != kill.VictimSide {
				stats.Assists++
			}
		}
	}

	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		sideKey := sideToString(playerSide)
		if sideKey == "" {
			continue
		}
		stats := sideStats[sideKey]

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
			if kill.KillerSteamID64 == player.SteamID64 {
				stats.FirstKills++
			}
			break
		}

		for _, kill := range killsInRound {
			if kill.IsVictimControllingBot || kill.IsSuicide() || kill.IsTeamKill() {
				continue
			}
			if kill.VictimSteamID64 == player.SteamID64 {
				stats.FirstDeaths++
			}
			break
		}
	}

	totalDamagePerSide := make(map[string]int)
	for _, damage := range match.Damages {
		if damage.AttackerSteamID64 != player.SteamID64 {
			continue
		}

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

	for sideKey, totalDamage := range totalDamagePerSide {
		if sideKey == "" {
			continue
		}
		if stats, ok := sideStats[sideKey]; ok && stats != nil && stats.RoundsPlayed > 0 {
			stats.ADR = float64(totalDamage) / float64(stats.RoundsPlayed)
		}
	}

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

// calculateKASTForSide calculates KAST percentage for a specific side.
// KAST = (Kill or Assist or Survived or Traded) / Total Rounds
func calculateKASTForSide(match *api.Match, player *api.Player, side common.Team) float64 {
	kastPerRound := make(map[int]bool)
	roundsOnThisSide := 0

	for _, round := range match.Rounds {
		playerSide := determinePlayerSideInRound(match, player, round)
		if playerSide != side {
			continue
		}

		roundsOnThisSide++
		kastPerRound[round.Number] = false
		playerSurvived := true

		for _, kill := range match.Kills {
			if round.Number != kill.RoundNumber {
				continue
			}

			isTeamKill := kill.KillerSide == kill.VictimSide
			if isTeamKill {
				continue
			}

			if kill.AssisterSteamID64 == player.SteamID64 {
				kastPerRound[round.Number] = true
			}

			if kill.KillerSteamID64 == player.SteamID64 && kill.VictimSteamID64 != player.SteamID64 {
				kastPerRound[round.Number] = true
			}

			if kill.VictimSteamID64 == player.SteamID64 {
				playerSurvived = false
				if kill.IsTradeDeath {
					kastPerRound[round.Number] = true
				}
			}
		}

		if playerSurvived {
			kastPerRound[round.Number] = true
		}
	}

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

// calculateOverallStats aggregates statistics across all maps and sides.
func calculateOverallStats(mapStats map[string]*MapStatistics) *OverallStatistics {
	overall := &OverallStatistics{}

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

	if overall.Deaths > 0 {
		overall.KD = float64(overall.Kills) / float64(overall.Deaths)
	} else if overall.Kills > 0 {
		overall.KD = float64(overall.Kills)
	}

	totalDamage := 0.0
	for _, mapStat := range mapStats {
		for _, sideStat := range mapStat.SideStats {
			totalDamage += sideStat.ADR * float64(sideStat.RoundsPlayed)
		}
	}
	if overall.RoundsPlayed > 0 {
		overall.ADR = totalDamage / float64(overall.RoundsPlayed)
	}

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
