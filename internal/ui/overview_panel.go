// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/nlamirault/e2c/internal/color"
)

// OverviewPanel represents the overview panel at the top of the UI
type OverviewPanel struct {
	ui               *UI
	view             *tview.TextView
	instanceCount    int
	region           string
	instancesRunning int
	instancesStopped int
	// Currently not using theme
}

// NewOverviewPanel creates a new overview panel
func NewOverviewPanel(ui *UI) *OverviewPanel {
	panel := &OverviewPanel{
		ui:   ui,
		view: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft),
	}

	// Set border and title with a more prominent style
	panel.view.SetBorder(true).
		SetTitle(" 🖥️  EC2 Dashboard ").
		SetBorderColor(color.AppColors.Border).
		SetTitleColor(color.AppColors.Title)

	// Set initial content
	panel.Update(0, 0, 0, "Unknown")

	return panel
}

// Update updates the overview panel content
func (p *OverviewPanel) Update(total, running, stopped int, region string) {
	p.instanceCount = total
	p.instancesRunning = running
	p.instancesStopped = stopped
	p.region = region

	// Calculate other instance states
	other := total - running - stopped

	// Use standard color names for simplicity
	headerColor := "yellow"
	runningColor := "green"
	stoppedColor := "red"
	otherColor := "yellow"
	regionColor := "blue"
	keyColor := "blue"
	textColor := "white"

	// Format the overview text
	text := fmt.Sprintf(`
 [::b][%s]EC2 INSTANCES[%s][::-]
 [%s]Total:[%s] %d     [%s]Running:[%s] %d     [%s]Stopped:[%s] %d     [%s]Other:[%s] %d

 [::b][%s]AWS REGION[%s][::-]
 [%s]%s[%s]

 [::b][%s]KEY MAPPINGS[%s][::-]
 [%s]?[%s]: Help       [%s]q[%s]: Quit       [%s]r[%s]: Refresh     [%s]f[%s]: Filter
 [%s]s[%s]: Start      [%s]p[%s]: Stop       [%s]b[%s]: Reboot      [%s]t[%s]: Terminate
 [%s]c[%s]: Connect    [%s]l[%s]: Logs       [%s]Esc[%s]: Back
`,
		headerColor, textColor,
		headerColor, textColor, p.instanceCount,
		runningColor, textColor, p.instancesRunning,
		stoppedColor, textColor, p.instancesStopped,
		otherColor, textColor, other,
		headerColor, textColor,
		regionColor, p.region, textColor,
		headerColor, textColor,
		keyColor, textColor, keyColor, textColor, keyColor, textColor, keyColor, textColor,
		keyColor, textColor, keyColor, textColor, keyColor, textColor, keyColor, textColor,
		keyColor, textColor, keyColor, textColor, keyColor, textColor,
	)

	p.view.SetText(text)
}

// UpdateStats updates just the instance statistics
func (p *OverviewPanel) UpdateStats(total, running, stopped int) {
	p.Update(total, running, stopped, p.region)
}

// UpdateRegion updates just the region information
func (p *OverviewPanel) UpdateRegion(region string) {
	p.Update(p.instanceCount, p.instancesRunning, p.instancesStopped, region)
}

// UpdateTheme updates the theme colors
func (p *OverviewPanel) UpdateTheme() {
	// Update border and title colors
	p.view.SetBorderColor(color.AppColors.Border)
	p.view.SetTitleColor(color.AppColors.Title)

	// Refresh the panel with new colors
	p.Update(p.instanceCount, p.instancesRunning, p.instancesStopped, p.region)
}

// getColorName maps a color to a standard name
// Keeping this as a stub for future color system improvements
func getColorName(c tcell.Color) string {
	return "white" // Default fallback
}
