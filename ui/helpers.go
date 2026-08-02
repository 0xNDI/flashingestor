package ui

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/Macmod/flashingestor/config"
)

// tviewTagPattern matches tview color/region tags such as [blue], [-],
// ["ingest"], [black:blue], etc. It is used to strip tags before printing
// log lines to stdout in headless mode.
var tviewTagPattern = regexp.MustCompile(`\[[^\]]*\]`)

func stripTviewTags(s string) string {
	return tviewTagPattern.ReplaceAllString(s, "")
}

// padString pads a string to the specified width for consistent table alignment
func padString(s string, width int) string {
	if len(s) < width {
		return fmt.Sprintf("%-*s", width, s)
	}
	return s
}

// SwitchToPage switches between the progress tracker pages
func (app *Application) SwitchToPage(pageName string) {
	app.currentPage = pageName
	app.progressPages.SwitchToPage(pageName)

	// Update page selector text with blue highlighting on active page
	switch pageName {
	case "ingest":
		app.pageSelector.SetText(`[blue]["ingest"](1) Ingest[""][-]  ["remote"](2) RemoteCollect[""]  ["conversion"](3) Convert[""]`)
	case "remote":
		app.pageSelector.SetText(`["ingest"](1) Ingest[""]  [blue]["remote"](2) RemoteCollect[""][-]  ["conversion"](3) Convert[""]`)
	case "conversion":
		app.pageSelector.SetText(`["ingest"](1) Ingest[""]  ["remote"](2) RemoteCollect[""]  [blue]["conversion"](3) Convert[""][-]`)
	}
}

// UpdateLog adds a message to the log panel with timestamp
func (app *Application) UpdateLog(message string) {
	if app.headless {
		// In headless mode there is no log panel; print to stdout instead.
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(os.Stdout, "[%s] %s\n", currentTime, stripTviewTags(message))
		return
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	// Called from logger goroutine
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(app.logPanel, "[white][%s][-] %s\n", formattedTime, message)
		app.logPanel.ScrollToEnd()
	})
}

// SetRuntimeOptions sets the runtime options reference for the UI
func (app *Application) SetRuntimeOptions(opts *config.RuntimeOptions) {
	app.runtimeOptions = opts
}
