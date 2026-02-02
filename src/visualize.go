package manalyzer

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// StartVisualizationServer starts HTTP server for visualization.
func StartVisualizationServer(data *WrangleResult) (string, error) {
	if data == nil {
		return "", fmt.Errorf("data is nil")
	}

	// Find available port
	var listener net.Listener
	var port int
	var err error
	
	for port = 8080; port <= 8090; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
	}
	
	if err != nil || listener == nil {
		return "", fmt.Errorf("no available ports in range 8080-8090: %w", err)
	}
	
	// Set up handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler(data))
	mux.HandleFunc("/player-comparison", playerComparisonHandler(data))
	mux.HandleFunc("/side-performance", sidePerformanceHandler(data))
	mux.HandleFunc("/map-breakdown", mapBreakdownHandler(data))
	
	// Start server in goroutine with panic recovery
	server := &http.Server{Handler: mux}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log panic but don't crash the main application
				fmt.Printf("PANIC in visualization server: %v\n", r)
			}
		}()
		server.Serve(listener)
	}()
	
	url := fmt.Sprintf("http://localhost:%d", port)
	return url, nil
}

// openBrowser opens URL in system browser.
func openBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("url is empty")
	}

	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "linux":
		browsers := []string{"xdg-open", "sensible-browser", "firefox", "chromium-browser", "google-chrome"}
		for _, browser := range browsers {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, url)
				return cmd.Start()
			}
		}
		return fmt.Errorf("no browser found")
		
	case "darwin":
		cmd = exec.Command("open", url)
		return cmd.Start()
		
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		return cmd.Start()
		
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// dashboardHandler serves the main dashboard page.
func dashboardHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %v", rec), http.StatusInternalServerError)
			}
		}()

		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Manalyzer - Player Statistics Dashboard</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background: #1a1a1a; color: #fff; }
		h1 { text-align: center; color: #4CAF50; }
		.chart-container { margin: 20px 0; }
		a { color: #4CAF50; text-decoration: none; margin: 0 15px; }
		a:hover { text-decoration: underline; }
		.nav { text-align: center; margin: 20px 0; padding: 15px; background: #2a2a2a; }
	</style>
</head>
<body>
	<h1>🎮 CS:GO Player Statistics Dashboard</h1>
	<div class="nav">
		<a href="/player-comparison">Player Comparison</a>
		<a href="/side-performance">T vs CT Performance</a>
		<a href="/map-breakdown">Map Breakdown</a>
	</div>
	<div class="chart-container">
		<h2>Quick Overview</h2>
		<p>Select a chart from the navigation above to view detailed statistics.</p>
		<p><strong>Total Matches Analyzed:</strong> ` + fmt.Sprintf("%d", data.TotalMatches) + `</p>
		<p><strong>Players Tracked:</strong> ` + fmt.Sprintf("%d", len(data.PlayerStats)) + `</p>
		<p><strong>Maps:</strong> ` + strings.Join(data.MapList, ", ") + `</p>
	</div>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

// playerComparisonHandler creates player comparison bar chart.
func playerComparisonHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %v", rec), http.StatusInternalServerError)
			}
		}()

		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "Player Comparison",
				Subtitle: "Overall Performance Metrics",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
			charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		var players []string
		kastData := make([]opts.BarData, 0)
		adrData := make([]opts.BarData, 0)
		kdData := make([]opts.BarData, 0)
		
		for _, player := range data.PlayerStats {
			if player != nil && player.OverallStats != nil {
				players = append(players, player.PlayerName)
				kastData = append(kastData, opts.BarData{Value: player.OverallStats.KAST})
				adrData = append(adrData, opts.BarData{Value: player.OverallStats.ADR})
				kdData = append(kdData, opts.BarData{Value: player.OverallStats.KD * 100})
			}
		}
		
		bar.SetXAxis(players)
		bar.AddSeries("KAST%", kastData)
		bar.AddSeries("ADR", adrData)
		bar.AddSeries("K/D x100", kdData)
		
		bar.Render(w)
	}
}

// sidePerformanceHandler creates T vs CT comparison chart.
func sidePerformanceHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %v", rec), http.StatusInternalServerError)
			}
		}()

		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "T vs CT Performance",
				Subtitle: "Side-Specific Statistics",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
			charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		var players []string
		tKAST := make([]opts.BarData, 0)
		ctKAST := make([]opts.BarData, 0)
		
		for _, player := range data.PlayerStats {
			if player != nil {
				players = append(players, player.PlayerName)
				
				var tKastSum, ctKastSum float64
				var tRounds, ctRounds int
				
				for _, mapStat := range player.MapStats {
					if tStats, ok := mapStat.SideStats["T"]; ok {
						tKastSum += (tStats.KAST / 100.0) * float64(tStats.RoundsPlayed)
						tRounds += tStats.RoundsPlayed
					}
					if ctStats, ok := mapStat.SideStats["CT"]; ok {
						ctKastSum += (ctStats.KAST / 100.0) * float64(ctStats.RoundsPlayed)
						ctRounds += ctStats.RoundsPlayed
					}
				}
				
				tAvg := 0.0
				if tRounds > 0 {
					tAvg = (tKastSum / float64(tRounds)) * 100.0
				}
				ctAvg := 0.0
				if ctRounds > 0 {
					ctAvg = (ctKastSum / float64(ctRounds)) * 100.0
				}
				
				tKAST = append(tKAST, opts.BarData{Value: tAvg})
				ctKAST = append(ctKAST, opts.BarData{Value: ctAvg})
			}
		}
		
		bar.SetXAxis(players)
		bar.AddSeries("T Side KAST%", tKAST)
		bar.AddSeries("CT Side KAST%", ctKAST)
		
		bar.Render(w)
	}
}

// mapBreakdownHandler creates map performance breakdown using bar chart.
func mapBreakdownHandler(data *WrangleResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %v", rec), http.StatusInternalServerError)
			}
		}()

		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "Map Performance Breakdown",
				Subtitle: "KAST% by Player and Map",
			}),
			charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
			charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		
		// Get unique maps
		mapSet := make(map[string]bool)
		for _, mapName := range data.MapList {
			mapSet[mapName] = true
		}
		maps := make([]string, 0, len(mapSet))
		for m := range mapSet {
			maps = append(maps, m)
		}
		
		bar.SetXAxis(maps)
		
		// Add series for each player
		for _, player := range data.PlayerStats {
			if player == nil {
				continue
			}
			
			playerData := make([]opts.BarData, 0)
			for _, mapName := range maps {
				kast := 0.0
				if mapStat, ok := player.MapStats[mapName]; ok {
					var kastSum float64
					var rounds int
					for _, sideStat := range mapStat.SideStats {
						if sideStat != nil {
							kastSum += (sideStat.KAST / 100.0) * float64(sideStat.RoundsPlayed)
							rounds += sideStat.RoundsPlayed
						}
					}
					if rounds > 0 {
						kast = (kastSum / float64(rounds)) * 100.0
					}
				}
				playerData = append(playerData, opts.BarData{Value: kast})
			}
			
			bar.AddSeries(player.PlayerName, playerData)
		}
		
		bar.Render(w)
	}
}
