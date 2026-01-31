package manalyzer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App    *tview.Application
	Pages  *tview.Pages
	Root   *tview.Flex
	Left   *tview.Box
	Middle tview.Primitive
	Bottom *tview.Box
}

func New() *UI {
	app := tview.NewApplication()

	left := tview.NewBox().SetBorder(true).SetTitle("Left (1/2 x width of Top)")
	middle := tview.NewTextView().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	bottom := tview.NewBox().SetBorder(true).SetTitle("Bottom (10 rows)")

	col := tview.NewFlex().
		AddItem(left, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(middle, 0, 3, false).
			AddItem(bottom, 10, 1, false), 0, 2, false)

	pages := tview.NewPages().AddPage("main", col, true, true)

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC, tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
	})

	return &UI{
		App:    app,
		Pages:  pages,
		Root:   col,
		Left:   left,
		Middle: middle,
		Bottom: bottom,
	}
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

func (u *UI) SetMiddleText(text string) {
	u.QueueUpdate(func() {
		if tv, ok := u.Middle.(*tview.TextView); ok {
			tv.Clear()
			tv.Write([]byte(text))
		}
	})
}
