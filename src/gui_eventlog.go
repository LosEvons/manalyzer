package manalyzer

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// EventLog displays timestamped event messages.
type EventLog struct {
	textView *tview.TextView
	maxLines int
	lines    []string
}

func newEventLog(maxLines int) *EventLog {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	
	tv.SetBorder(true)
	tv.SetTitle("Event Log")

	tv.SetChangedFunc(func() {
		tv.ScrollToEnd()
	})

	return &EventLog{
		textView: tv,
		maxLines: maxLines,
		lines:    make([]string, 0, maxLines),
	}
}

func (el *EventLog) Log(message string) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[yellow]%s[-] %s", timestamp, message)

	el.lines = append(el.lines, line)

	if len(el.lines) > el.maxLines {
		el.lines = el.lines[len(el.lines)-el.maxLines:]
	}

	var builder strings.Builder
	for i, l := range el.lines {
		builder.WriteString(l)
		if i < len(el.lines)-1 {
			builder.WriteString("\n")
		}
	}
	el.textView.SetText(builder.String())
}

func (el *EventLog) LogError(message string) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[yellow]%s[-] [red]ERROR:[-] %s", timestamp, message)

	el.lines = append(el.lines, line)
	if len(el.lines) > el.maxLines {
		el.lines = el.lines[len(el.lines)-el.maxLines:]
	}

	var builder strings.Builder
	for i, l := range el.lines {
		builder.WriteString(l)
		if i < len(el.lines)-1 {
			builder.WriteString("\n")
		}
	}
	el.textView.SetText(builder.String())
}
