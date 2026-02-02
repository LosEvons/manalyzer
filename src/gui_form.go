package manalyzer

import (
	"fmt"
	"os"
	"time"

	"github.com/rivo/tview"
)

// PlayerInput represents user input for player tracking.
type PlayerInput struct {
	Name      string
	SteamID64 string
}

// AnalysisConfig holds configuration for analysis.
type AnalysisConfig struct {
	Players  [5]PlayerInput
	BasePath string
}

func createPlayerInputForm() *tview.Form {
	return createPlayerInputFormWithConfig(DefaultConfig())
}

func createPlayerInputFormWithConfig(config *Config) *tview.Form {
	form := tview.NewForm()
	
	form.SetBorder(true)
	form.SetTitle("Player Configuration")
	form.SetTitleAlign(tview.AlignLeft)

	for i := 0; i < 5; i++ {
		playerName := ""
		steamID := ""
		if i < len(config.Players) {
			playerName = config.Players[i].Name
			steamID = config.Players[i].SteamID64
		}
		
		playerLabel := fmt.Sprintf("Player %d Name", i+1)
		steamLabel := fmt.Sprintf("Player %d SteamID64", i+1)

		form.AddInputField(playerLabel, playerName, 30, nil, nil)
		form.AddInputField(steamLabel, steamID, 17, validateSteamID64, nil)
	}

	form.AddInputField("Demo Base Path", config.LastDemoPath, 50, nil, nil)

	form.AddButton("Analyze", nil)
	form.AddButton("Clear", nil)
	form.AddButton("Save Config", nil)
	form.AddButton("Visualize", nil)

	return form
}

func validateSteamID64(text string, lastChar rune) bool {
	if text == "" {
		return true
	}
	if lastChar < '0' || lastChar > '9' {
		return false
	}
	return len(text) <= 17
}

func (u *UI) setupFormHandlers(form *tview.Form) {
	analyzeIdx := form.GetButtonCount() - 4
	clearIdx := form.GetButtonCount() - 3
	saveIdx := form.GetButtonCount() - 2
	visualizeIdx := form.GetButtonCount() - 1

	form.GetButton(analyzeIdx).SetSelectedFunc(func() {
		u.onAnalyzeClicked(form)
	})

	form.GetButton(clearIdx).SetSelectedFunc(func() {
		u.onClearClicked(form)
	})

	form.GetButton(saveIdx).SetSelectedFunc(func() {
		u.onSaveConfigClicked(form)
	})

	form.GetButton(visualizeIdx).SetSelectedFunc(func() {
		u.onVisualizeClicked()
	})
}

func (u *UI) onAnalyzeClicked(form *tview.Form) {
	config := u.extractConfigFromForm(form)

	if config.BasePath == "" {
		u.logError("Demo base path must be specified")
		return
	}

	if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
		u.logError(fmt.Sprintf("Path does not exist: %s", config.BasePath))
		return
	}

	validPlayers := 0
	for _, player := range config.Players {
		if player.SteamID64 != "" {
			validPlayers++
		}
	}

	if validPlayers == 0 {
		u.logError("At least one player with SteamID64 must be specified")
		return
	}

	// Save config if auto-save enabled
	if u.config.Preferences.AutoSave {
		persistConfig := u.buildConfigFromForm(form)
		if err := SaveConfig(persistConfig); err != nil {
			u.logEvent(fmt.Sprintf("Warning: Failed to save config: %v", err))
		} else {
			u.config = persistConfig
		}
	}

	go u.runAnalysis(config)
}

func (u *UI) onClearClicked(form *tview.Form) {
	formItemCount := form.GetFormItemCount()
	for i := 0; i < formItemCount; i++ {
		if field, ok := form.GetFormItem(i).(*tview.InputField); ok {
			field.SetText("")
		}
	}
	u.logEvent("Form cleared")
}

func (u *UI) onSaveConfigClicked(form *tview.Form) {
	config := u.buildConfigFromForm(form)
	if err := SaveConfig(config); err != nil {
		u.logError(fmt.Sprintf("Failed to save config: %v", err))
	} else {
		u.config = config
		u.logEvent("Configuration saved successfully")
	}
}

func (u *UI) onVisualizeClicked() {
	if u.statsTable.data == nil {
		u.logError("No data to visualize. Run analysis first.")
		return
	}
	
	u.logEvent("Starting visualization server...")
	
	url, err := StartVisualizationServer(u.statsTable.data)
	if err != nil {
		u.logError(fmt.Sprintf("Failed to start server: %v", err))
		return
	}
	
	u.logEvent(fmt.Sprintf("Visualization server started at %s", url))
	
	time.Sleep(500 * time.Millisecond)
	
	if err := openBrowser(url); err != nil {
		u.logEvent(fmt.Sprintf("Could not open browser: %v", err))
		u.logEvent(fmt.Sprintf("Visit %s manually to view charts", url))
	} else {
		u.logEvent("Visualization dashboard opened in browser")
	}
}

func (u *UI) extractConfigFromForm(form *tview.Form) AnalysisConfig {
	config := AnalysisConfig{}

	for i := 0; i < 5; i++ {
		nameIdx := i * 2
		steamIdx := i*2 + 1

		if nameField, ok := form.GetFormItem(nameIdx).(*tview.InputField); ok {
			config.Players[i].Name = nameField.GetText()
		}
		if steamField, ok := form.GetFormItem(steamIdx).(*tview.InputField); ok {
			config.Players[i].SteamID64 = steamField.GetText()
		}
	}

	if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
		config.BasePath = pathField.GetText()
	}

	return config
}

func (u *UI) buildConfigFromForm(form *tview.Form) *Config {
	config := &Config{
		Version: 1,
		Players: make([]PlayerConfig, 0),
		Preferences: Preferences{
			AutoSave: true,
		},
	}
	
	for i := 0; i < 5; i++ {
		nameIdx := i * 2
		steamIdx := i*2 + 1

		var name, steamID string
		if nameField, ok := form.GetFormItem(nameIdx).(*tview.InputField); ok {
			name = nameField.GetText()
		}
		if steamField, ok := form.GetFormItem(steamIdx).(*tview.InputField); ok {
			steamID = steamField.GetText()
		}
		
		if steamID != "" {
			config.Players = append(config.Players, PlayerConfig{
				Name:      name,
				SteamID64: steamID,
			})
		}
	}

	if pathField, ok := form.GetFormItem(10).(*tview.InputField); ok {
		config.LastDemoPath = pathField.GetText()
	}

	return config
}

func (u *UI) runAnalysis(config AnalysisConfig) {
	defer func() {
		if r := recover(); r != nil {
			u.logError(fmt.Sprintf("PANIC during analysis: %v", r))
		}
	}()
	
	u.logEvent("Starting analysis...")

	var steamIDs []string
	for _, player := range config.Players {
		if player.SteamID64 != "" {
			steamIDs = append(steamIDs, player.SteamID64)
			u.logEvent(fmt.Sprintf("Tracking player: %s (%s)", player.Name, player.SteamID64))
		}
	}

	u.logEvent(fmt.Sprintf("Searching for demos in: %s", config.BasePath))
	matches, err := GatherAllDemosFromPath(config.BasePath)

	if err != nil {
		if len(matches) == 0 {
			u.logError(fmt.Sprintf("%v", err))
			return
		}
		u.logEvent(fmt.Sprintf("Warning during demo gathering: %v", err))
	}

	if len(matches) == 0 {
		u.logError("No demo files found or all demos failed to parse")
		return
	}

	u.logEvent(fmt.Sprintf("Found %d demos, starting analysis...", len(matches)))

	result, err := ProcessMatches(matches, steamIDs)
	if err != nil {
		u.logError(fmt.Sprintf("Error during analysis: %v", err))
		return
	}

	u.logEvent(fmt.Sprintf("Analysis complete! Processed %d matches", result.TotalMatches))
	u.logEvent(fmt.Sprintf("Found stats for %d players across %d maps",
		len(result.PlayerStats), len(result.MapList)))

	u.QueueUpdate(func() {
		u.statsTable.UpdateData(result)
	})
}
