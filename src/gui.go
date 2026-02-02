package manalyzer

import (
"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

const (
eventLogHeight = 5
)

// UI manages the terminal user interface.
type UI struct {
App        *tview.Application
Pages      *tview.Pages
Root       *tview.Flex
form       *tview.Form
eventLog   *EventLog
statsTable *StatisticsTable
config     *Config // Store loaded configuration
}

func New() *UI {
app := tview.NewApplication()

// Load config
config, err := LoadConfig()
if err != nil {
config = DefaultConfig()
}

// Create components
form := createPlayerInputFormWithConfig(config)
eventLog := newEventLog(50)
statsTable := newStatisticsTable(app) // Pass app reference

// Create layout
leftPanel := form

middlePanel := eventLog.textView
middlePanel.SetBorder(true).
SetTitle("Event Log").
SetTitleAlign(tview.AlignLeft)

bottomPanel := statsTable.table

rightColumn := tview.NewFlex().
SetDirection(tview.FlexRow).
AddItem(middlePanel, eventLogHeight, 0, false).
AddItem(bottomPanel, 0, 1, false)

mainLayout := tview.NewFlex().
AddItem(leftPanel, 0, 1, true).
AddItem(rightColumn, 0, 2, false)

pages := tview.NewPages().AddPage("main", mainLayout, true, true)

app.SetRoot(pages, true).EnableMouse(true)
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyESC, tcell.KeyCtrlC:
app.Stop()
return nil
}
return event
})

ui := &UI{
App:        app,
Pages:      pages,
Root:       mainLayout,
form:       form,
eventLog:   eventLog,
statsTable: statsTable,
config:     config,
}

ui.setupFormHandlers(form)

return ui
}

func (u *UI) Start() error {
return u.App.Run()
}

func (u *UI) Stop() {
u.App.Stop()
}

func (u *UI) QueueUpdate(fn func()) {
u.App.QueueUpdateDraw(fn)
}

func (u *UI) logEvent(message string) {
u.QueueUpdate(func() {
u.eventLog.Log(message)
})
}

func (u *UI) logError(message string) {
u.QueueUpdate(func() {
u.eventLog.LogError(message)
})
}
