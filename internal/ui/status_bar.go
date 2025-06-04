// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/nlamirault/e2c/internal/color"
)

// StatusBar represents the status bar at the bottom of the UI
type StatusBar struct {
	ui       *UI
	view     *tview.TextView
	status   string
	region   string
	lastSync time.Time
	mode     string // Current UI mode
}

// NewStatusBar creates a new status bar
func NewStatusBar(ui *UI) *StatusBar {
	bar := &StatusBar{
		ui:     ui,
		view:   tview.NewTextView().SetDynamicColors(true),
		status: "Starting...",
		mode:   "normal",
	}

	// Set background color from theme
	bar.view.SetBackgroundColor(color.AppColors.Background)

	// Update the view
	bar.update()

	return bar
}

// UpdateTheme updates the status bar theme
func (b *StatusBar) UpdateTheme() {
	// Update the background color based on theme
	b.view.SetBackgroundColor(color.AppColors.Background)
	b.update()
}

// SetStatus sets the status message
func (b *StatusBar) SetStatus(status string) {
	b.status = status
	b.lastSync = time.Now()
	b.update()
}

// SetError sets an error message in the status bar
func (b *StatusBar) SetError(err string) {
	// Use standard color name for simplicity
	b.status = fmt.Sprintf("[red]%s[-]", err)
	b.update()
}

// SetRegion sets the current region
func (b *StatusBar) SetRegion(region string) {
	b.region = region
	b.update()
}

// SetMode sets the current UI mode
func (b *StatusBar) SetMode(mode string) {
	b.mode = mode
	b.update()
}

// update updates the status bar content
func (b *StatusBar) update() {
	// Use standard color names for simplicity
	labelColor := "yellow"
	valueColor := "white"
	modeValueColor := "blue"

	var regionInfo string
	if b.region != "" {
		regionInfo = fmt.Sprintf("[%s]Region:[%s] %s", labelColor, valueColor, b.region)
	}

	var lastSyncInfo string
	if !b.lastSync.IsZero() {
		lastSyncInfo = fmt.Sprintf("[%s]Last sync:[%s] %s", labelColor, valueColor, b.lastSync.Format("15:04:05"))
	}

	var modeInfo string
	switch b.mode {
	case "filtering":
		modeInfo = fmt.Sprintf("[%s]Mode:[%s] [%s]Filtering[%s]", labelColor, valueColor, modeValueColor, valueColor)
	case "selecting":
		modeInfo = fmt.Sprintf("[%s]Mode:[%s] [%s]Selecting[%s]", labelColor, valueColor, modeValueColor, valueColor)
	case "normal":
		modeInfo = fmt.Sprintf("[%s]Mode:[%s] [%s]Normal[%s]", labelColor, valueColor, modeValueColor, valueColor)
	}

	status := b.status
	if status == "" {
		status = "Ready"
	}

	// Build status text with all components
	components := []string{status}

	if regionInfo != "" {
		components = append(components, regionInfo)
	}

	if modeInfo != "" {
		components = append(components, modeInfo)
	}

	if lastSyncInfo != "" {
		components = append(components, lastSyncInfo)
	}

	// Join all components with a separator
	text := " " + strings.Join(components, " | ") + " "

	b.view.SetText(text)
}

// Clear clears the status message
func (b *StatusBar) Clear() {
	b.status = ""
	b.update()
}
