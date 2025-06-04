// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package color

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Colors represents the colors used in the application
type Colors struct {
	// Basic colors
	Background tcell.Color
	Foreground tcell.Color
	Border     tcell.Color
	Title      tcell.Color
	Selected   tcell.Color

	// UI element colors
	HeaderFg tcell.Color
	HeaderBg tcell.Color

	// Status colors
	Running tcell.Color
	Stopped tcell.Color
	Pending tcell.Color
	Error   tcell.Color

	// Other colors
	Highlight tcell.Color
	Secondary tcell.Color
}

// Global color scheme that uses Nord theme colors
var AppColors = Colors{
	Background: tcell.GetColor("#2E3440"), // Primary background
	Foreground: tcell.GetColor("#D8DEE9"), // Primary foreground
	Border:     tcell.GetColor("#81A1C1"), // Normal blue
	Title:      tcell.GetColor("#88C0D0"), // Normal cyan
	Selected:   tcell.GetColor("#3B4252"), // Normal black
	HeaderFg:   tcell.GetColor("#ECEFF4"), // Bright white
	HeaderBg:   tcell.GetColor("#4C566A"), // Bright black
	Running:    tcell.GetColor("#A3BE8C"), // Normal green
	Stopped:    tcell.GetColor("#BF616A"), // Normal red
	Pending:    tcell.GetColor("#EBCB8B"), // Normal yellow
	Error:      tcell.GetColor("#BF616A"), // Normal red
	Highlight:  tcell.GetColor("#EBCB8B"), // Normal yellow
	Secondary:  tcell.GetColor("#81A1C1"), // Normal blue
}

// InitializeColors applies the application colors to tview components
func InitializeColors() {
	// Apply colors to tview global styles
	tview.Styles.PrimitiveBackgroundColor = AppColors.Background
	tview.Styles.ContrastBackgroundColor = AppColors.HeaderBg
	tview.Styles.MoreContrastBackgroundColor = AppColors.Selected
	tview.Styles.BorderColor = AppColors.Border
	tview.Styles.TitleColor = AppColors.Title
	tview.Styles.GraphicsColor = AppColors.Border
	tview.Styles.PrimaryTextColor = AppColors.Foreground
	tview.Styles.SecondaryTextColor = AppColors.Secondary
	tview.Styles.TertiaryTextColor = AppColors.Highlight
	tview.Styles.InverseTextColor = AppColors.HeaderFg
	tview.Styles.ContrastSecondaryTextColor = AppColors.Highlight
}
