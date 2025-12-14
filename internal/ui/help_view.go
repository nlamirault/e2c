// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/nlamirault/e2c/internal/color"
)

// HelpView represents the help bar at the bottom of the UI
type HelpView struct {
	view       *tview.TextView
	expertMode bool
}

// NewHelpView creates a new help view
func NewHelpView(expertMode bool) *HelpView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Set background color from theme
	view.SetBackgroundColor(color.AppColors.HeaderBg)

	// Update help text
	helpText := "[yellow]?[white]:Help  [yellow]q[white]:Quit  [yellow]r[white]:Refresh  [yellow]f[white]:Filter  [yellow]s[white]:Start  [yellow]p[white]:Stop  [yellow]b[white]:Reboot  [yellow]t[white]:Terminate  [yellow]c[white]:Connect  [yellow]l[white]:Logs"
	if expertMode {
		helpText += "  [yellow]x[white]:Term.Protect  [yellow]n[white]:Stop.Protect"
	}

	view.SetText(helpText)

	return &HelpView{
		view:       view,
		expertMode: expertMode,
	}
}

// SetText sets the help text
func (h *HelpView) SetText(text string) {
	h.view.SetText(text)
	// Update the background color when text changes
	h.view.SetBackgroundColor(color.AppColors.HeaderBg)
}

// Clear clears the help text
func (h *HelpView) Clear() {
	h.view.SetText("")
}

// Update updates the help text based on context
func (h *HelpView) Update(context string) {
	// Use standard color names for simplicity
	highlightColor := "yellow"
	textColor := "white"

	switch context {
	case "main":
		mainText := fmt.Sprintf("[%s]?[%s]:Help [%s]q[%s]:Quit [%s]r[%s]:Refresh [%s]f[%s]:Filter [%s]s[%s]:Start [%s]p[%s]:Stop [%s]b[%s]:Reboot [%s]t[%s]:Terminate [%s]c[%s]:Connect [%s]l[%s]:Logs",
			highlightColor, textColor, highlightColor, textColor, highlightColor, textColor, highlightColor, textColor,
			highlightColor, textColor, highlightColor, textColor, highlightColor, textColor, highlightColor, textColor,
			highlightColor, textColor, highlightColor, textColor)
		if h.expertMode {
			mainText += fmt.Sprintf(" [%s]x[%s]:Term.Protect [%s]n[%s]:Stop.Protect", highlightColor, textColor, highlightColor, textColor)
		}
		h.view.SetText(mainText)
	case "detail":
		detailText := fmt.Sprintf("[%s]Esc[%s]:Back [%s]s[%s]:Start [%s]p[%s]:Stop [%s]b[%s]:Reboot [%s]t[%s]:Terminate [%s]c[%s]:Connect [%s]l[%s]:Logs",
			highlightColor, textColor, highlightColor, textColor, highlightColor, textColor, highlightColor, textColor,
			highlightColor, textColor, highlightColor, textColor, highlightColor, textColor)
		if h.expertMode {
			detailText += fmt.Sprintf(" [%s]x[%s]:Term.Protect [%s]n[%s]:Stop.Protect", highlightColor, textColor, highlightColor, textColor)
		}
		h.view.SetText(detailText)
	case "modal":
		h.view.SetText(fmt.Sprintf("[%s]Esc[%s]:Close", highlightColor, textColor))
	default:
		h.view.SetText(fmt.Sprintf("[%s]?[%s]:Help [%s]q[%s]:Quit", highlightColor, textColor, highlightColor, textColor))
	}

	// Update the background color
	h.view.SetBackgroundColor(color.AppColors.HeaderBg)
}
