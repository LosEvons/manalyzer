package manalyzer

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatisticsTable displays player statistics.
type StatisticsTable struct {
	table      *tview.Table
	data       *WrangleResult
	filterMap  string
	filterSide string
	app        *tview.Application  // NEW: For UI updates
	sortColumn int                 // NEW: Current sort column (0-11)
	sortDesc   bool                // NEW: Sort direction
}

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

func (st *StatisticsTable) UpdateData(result *WrangleResult) {
	st.data = result
	st.renderTable()
}

func (st *StatisticsTable) SetFilter(mapFilter, sideFilter string) {
	st.filterMap = mapFilter
	st.filterSide = sideFilter
	st.renderTable()
}

func (st *StatisticsTable) toggleSort(columnIndex int) {
	if st.sortColumn == columnIndex {
		st.sortDesc = !st.sortDesc
	} else {
		st.sortColumn = columnIndex
		if columnIndex == 0 || columnIndex == 1 {
			st.sortDesc = false  // Ascending for names
		} else {
			st.sortDesc = true   // Descending for stats
		}
	}
}

func (st *StatisticsTable) sortData(players []*PlayerStats) []*PlayerStats {
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
		case 8: // FK
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.FirstKills < sorted[j].OverallStats.FirstKills
			}
		case 9: // FD
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.FirstDeaths < sorted[j].OverallStats.FirstDeaths
			}
		case 10: // TK
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.TradeKills < sorted[j].OverallStats.TradeKills
			}
		case 11: // TD
			if sorted[i].OverallStats != nil && sorted[j].OverallStats != nil {
				less = sorted[i].OverallStats.TradeDeaths < sorted[j].OverallStats.TradeDeaths
			}
		default:
			less = sorted[i].PlayerName < sorted[j].PlayerName
		}
		
		if st.sortDesc {
			return !less
		}
		return less
	})
	
	return sorted
}

func (st *StatisticsTable) renderTable() {
	st.table.Clear()

	// Header row
	headers := []string{"Player", "Map", "Side", "KAST%", "ADR", "K/D",
		"Kills", "Deaths", "FK", "FD", "TK", "TD"}

	for col, header := range headers {
		columnIndex := col  // Capture loop variable
		
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
		
		// Make stat columns clickable
		if columnIndex >= 3 {
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

	// Data rows
	row := 1
	if st.data != nil {
		sortedPlayers := make([]*PlayerStats, 0, len(st.data.PlayerStats))
		
		for _, ps := range st.data.PlayerStats {
			if ps != nil {
				sortedPlayers = append(sortedPlayers, ps)
			}
		}
		
		sortedPlayers = st.sortData(sortedPlayers)

		for _, playerStats := range sortedPlayers {
			if playerStats == nil {
				continue
			}
			
			for mapName, mapStats := range playerStats.MapStats {
				if st.filterMap != "" && mapName != st.filterMap {
					continue
				}

				for _, side := range []string{"T", "CT"} {
					if st.filterSide != "" && side != st.filterSide {
						continue
					}

					if sideStats, ok := mapStats.SideStats[side]; ok {
						st.addDataRow(row, playerStats.PlayerName, mapName, side, sideStats)
						row++
					}
				}
				
				if st.filterSide == "" {
					st.addMapSummaryRow(row, playerStats.PlayerName, mapName, mapStats)
					row++
				}
			}

			if st.filterMap == "" && st.filterSide == "" && playerStats.OverallStats != nil {
				st.addOverallRow(row, playerStats.PlayerName, playerStats.OverallStats)
				row++
			}
		}
	}
}

func (st *StatisticsTable) addDataRow(row int, playerName, mapName, side string, stats *SideStatistics) {
	if stats == nil {
		return
	}
	
	cols := []string{
		playerName,
		mapName,
		side,
		fmt.Sprintf("%.1f", stats.KAST),
		fmt.Sprintf("%.1f", stats.ADR),
		fmt.Sprintf("%.2f", stats.KD),
		fmt.Sprintf("%d", stats.Kills),
		fmt.Sprintf("%d", stats.Deaths),
		fmt.Sprintf("%d", stats.FirstKills),
		fmt.Sprintf("%d", stats.FirstDeaths),
		fmt.Sprintf("%d", stats.TradeKills),
		fmt.Sprintf("%d", stats.TradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorWhite)
		st.table.SetCell(row, col, cell)
	}
}

func (st *StatisticsTable) addMapSummaryRow(row int, playerName, mapName string, mapStats *MapStatistics) {
	if mapStats == nil || mapStats.SideStats == nil {
		return
	}
	
	var totalKills, totalDeaths, totalAssists int
	var totalFirstKills, totalFirstDeaths int
	var totalTradeKills, totalTradeDeaths int
	var totalHeadshots, totalRoundsPlayed int
	var weightedKAST, weightedADR float64
	
	for _, sideStats := range mapStats.SideStats {
		if sideStats == nil {
			continue
		}
		totalKills += sideStats.Kills
		totalDeaths += sideStats.Deaths
		totalAssists += sideStats.Assists
		totalFirstKills += sideStats.FirstKills
		totalFirstDeaths += sideStats.FirstDeaths
		totalTradeKills += sideStats.TradeKills
		totalTradeDeaths += sideStats.TradeDeaths
		totalHeadshots += sideStats.Headshots
		totalRoundsPlayed += sideStats.RoundsPlayed
		
		weightedKAST += (sideStats.KAST / 100.0) * float64(sideStats.RoundsPlayed)
		weightedADR += sideStats.ADR * float64(sideStats.RoundsPlayed)
	}
	
	kast := 0.0
	adr := 0.0
	if totalRoundsPlayed > 0 {
		kast = (weightedKAST / float64(totalRoundsPlayed)) * 100.0
		adr = weightedADR / float64(totalRoundsPlayed)
	}
	
	kd := 0.0
	if totalDeaths > 0 {
		kd = float64(totalKills) / float64(totalDeaths)
	} else if totalKills > 0 {
		kd = float64(totalKills)
	}
	
	cols := []string{
		playerName,
		mapName,
		"Both",
		fmt.Sprintf("%.1f", kast),
		fmt.Sprintf("%.1f", adr),
		fmt.Sprintf("%.2f", kd),
		fmt.Sprintf("%d", totalKills),
		fmt.Sprintf("%d", totalDeaths),
		fmt.Sprintf("%d", totalFirstKills),
		fmt.Sprintf("%d", totalFirstDeaths),
		fmt.Sprintf("%d", totalTradeKills),
		fmt.Sprintf("%d", totalTradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorAqua).
			SetAttributes(tcell.AttrBold)
		st.table.SetCell(row, col, cell)
	}
}

func (st *StatisticsTable) addOverallRow(row int, playerName string, stats *OverallStatistics) {
	if stats == nil {
		return
	}
	
	cols := []string{
		playerName,
		"Overall",
		"All",
		fmt.Sprintf("%.1f", stats.KAST),
		fmt.Sprintf("%.1f", stats.ADR),
		fmt.Sprintf("%.2f", stats.KD),
		fmt.Sprintf("%d", stats.Kills),
		fmt.Sprintf("%d", stats.Deaths),
		fmt.Sprintf("%d", stats.FirstKills),
		fmt.Sprintf("%d", stats.FirstDeaths),
		fmt.Sprintf("%d", stats.TradeKills),
		fmt.Sprintf("%d", stats.TradeDeaths),
	}

	for col, text := range cols {
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorGreen).
			SetAttributes(tcell.AttrBold)
		st.table.SetCell(row, col, cell)
	}
}
