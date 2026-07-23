// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/nlamirault/e2c/internal/color"
	"github.com/nlamirault/e2c/internal/model"
)

// InstancesView represents the instances table view
type InstancesView struct {
	ui              *UI
	table           *tview.Table
	instances       []model.Instance
	instancesM      sync.Mutex
	selected        int
	headers         []string
	headerColor     tcell.Color
	textColor       tcell.Color
	tagColor        tcell.Color
	runningColor    tcell.Color
	stoppedColor    tcell.Color
	pendingColor    tcell.Color
	showProtections bool
	// Theme support will be added in future versions
}

// NewInstancesView creates a new instances view
func NewInstancesView(ui *UI) *InstancesView {
	v := &InstancesView{
		ui:              ui,
		table:           tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		instances:       make([]model.Instance, 0),
		selected:        0,
		headers:         []string{"ID", "Name", "State", "Type", "Region", "Private IP", "Public IP", "Age"},
		headerColor:     color.AppColors.Title,
		textColor:       color.AppColors.Foreground,
		tagColor:        color.AppColors.Secondary,
		runningColor:    color.AppColors.Running,
		stoppedColor:    color.AppColors.Stopped,
		pendingColor:    color.AppColors.Pending,
		showProtections: ui.config.UI.ExpertMode,
	}

	if v.showProtections {
		v.headers = append(v.headers, "T.Protect", "S.Protect")
	}

	// Set up table
	v.table.SetBorder(true).
		SetTitle("EC2 Instances").
		SetBorderColor(color.AppColors.Border).
		SetTitleColor(color.AppColors.Title)

	// Set up cell selection handler
	v.table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row-1 < len(v.instances) {
			v.ShowInstanceDetails(v.instances[row-1])
		}
	})

	// Return instance view
	return v
}

// UpdateInstances updates the instances table with new data
func (v *InstancesView) UpdateInstances(instances []model.Instance) {
	v.instancesM.Lock()
	defer v.instancesM.Unlock()

	v.instances = instances
	v.table.Clear()

	// Add headers
	// Set headers
	for i, header := range v.headers {
		v.table.SetCell(0, i,
			tview.NewTableCell(" "+header+" ").
				SetTextColor(v.headerColor).
				SetSelectable(false).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold).
				SetBackgroundColor(color.AppColors.HeaderBg))
	}

	// Add instances
	for i, instance := range instances {
		row := i + 1
		stateColor := getStateColor(instance.State)

		// Set ID
		v.table.SetCell(row, 0,
			tview.NewTableCell(" "+instance.ID+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set Name
		v.table.SetCell(row, 1,
			tview.NewTableCell(" "+instance.Name+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set State with color and emoji before state name
		v.table.SetCell(row, 2,
			tview.NewTableCell(" "+getStateEmoji(instance.State)+" "+instance.State+" ").
				SetTextColor(stateColor).
				SetAlign(tview.AlignLeft))

		// Set Type
		v.table.SetCell(row, 3,
			tview.NewTableCell(" "+instance.Type+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set Region
		v.table.SetCell(row, 4,
			tview.NewTableCell(" "+instance.Region+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set Private IP
		v.table.SetCell(row, 5,
			tview.NewTableCell(" "+instance.PrivateIP+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set Public IP
		v.table.SetCell(row, 6,
			tview.NewTableCell(" "+instance.PublicIP+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignLeft))

		// Set Age
		v.table.SetCell(row, 7,
			tview.NewTableCell(" "+formatDuration(instance.Age)+" ").
				SetTextColor(v.textColor).
				SetAlign(tview.AlignRight))

		if v.showProtections {
			protectionText := formatProtectionCell(instance.TerminationProtection, instance.TerminationProtectionKnown)
			v.table.SetCell(row, 8,
				tview.NewTableCell(" "+protectionText+" ").
					SetTextColor(v.textColor).
					SetAlign(tview.AlignCenter))

			stopProtectionText := formatProtectionCell(instance.StopProtection, instance.StopProtectionKnown)
			v.table.SetCell(row, 9,
				tview.NewTableCell(" "+stopProtectionText+" ").
					SetTextColor(v.textColor).
					SetAlign(tview.AlignCenter))
		}
	}

	// Restore selection if possible
	if v.selected < len(instances) {
		v.table.Select(v.selected+1, 0)
	} else if len(instances) > 0 {
		v.table.Select(1, 0)
		v.selected = 0
	}
}

// GetSelectedInstance returns the currently selected instance
func (v *InstancesView) GetSelectedInstance() *model.Instance {
	v.instancesM.Lock()
	defer v.instancesM.Unlock()

	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.instances) {
		return nil
	}

	v.selected = row - 1

	// Highlight the selected row is handled by tview automatically

	return &v.instances[v.selected]
}

// UpdateProtection updates the cached protection values for an instance and refreshes visible cells.
func (v *InstancesView) UpdateProtection(instanceID string, terminationProtection, stopProtection bool) {
	v.instancesM.Lock()
	defer v.instancesM.Unlock()

	var rowIndex int
	found := false
	for idx, inst := range v.instances {
		if inst.ID == instanceID {
			v.instances[idx].TerminationProtection = terminationProtection
			v.instances[idx].StopProtection = stopProtection
			v.instances[idx].TerminationProtectionKnown = true
			v.instances[idx].StopProtectionKnown = true
			rowIndex = idx + 1
			found = true
			break
		}
	}

	if !found || !v.showProtections {
		return
	}

	terminationText := formatProtectionCell(terminationProtection, true)
	v.table.SetCell(rowIndex, 8,
		tview.NewTableCell(" "+terminationText+" ").
			SetTextColor(v.textColor).
			SetAlign(tview.AlignCenter))

	stopText := formatProtectionCell(stopProtection, true)
	v.table.SetCell(rowIndex, 9,
		tview.NewTableCell(" "+stopText+" ").
			SetTextColor(v.textColor).
			SetAlign(tview.AlignCenter))
}

// ShowInstanceDetails displays a detailed view of an instance
func (v *InstancesView) ShowInstanceDetails(instance model.Instance) {
	inst := instance

	if term, stop, ok := v.ui.ec2Client.GetCachedProtectionStatus(inst.ID); ok {
		inst.TerminationProtection = term
		inst.StopProtection = stop
		inst.TerminationProtectionKnown = true
		inst.StopProtectionKnown = true
	} else {
		termProtect, stopProtect, err := v.ui.ec2Client.RefreshProtectionStatus(v.ui.ctx, inst.ID)
		if err != nil {
			v.ui.statusBar.SetError(fmt.Sprintf("Failed to load protections: %v", err))
		} else {
			inst.TerminationProtection = termProtect
			inst.StopProtection = stopProtect
			inst.TerminationProtectionKnown = true
			inst.StopProtectionKnown = true
			v.UpdateProtection(inst.ID, termProtect, stopProtect)
		}
	}

	detailsText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetScrollable(true).
		SetWrap(true)

		// Format instance details
	baseDetails := fmt.Sprintf(`
[::b][yellow]Instance Details[white][::-]
  [blue]ID:[white]            %s
  [blue]Name:[white]          %s
  [blue]Type:[white]          %s
  [blue]State:[white]         %s %s
  [blue]Region:[white]        %s
  [blue]Launch Time:[white]   %s
  [blue]Age:[white]           %s
  [blue]Private IP:[white]    %s
  [blue]Public IP:[white]     %s
  [blue]Platform:[white]      %s
  [blue]Architecture:[white]  %s
  [blue]T.Protect:[white]     %s
  [blue]S.Protect:[white]     %s
`,
		inst.ID,
		inst.Name,
		inst.Type,
		getStateEmoji(inst.State), inst.State,
		inst.Region,
		inst.LaunchTime.Format("2006-01-02 15:04:05"),
		formatDuration(inst.Age),
		inst.PrivateIP,
		inst.PublicIP,
		inst.Platform,
		inst.Architecture,
		formatProtectionStatus(inst.TerminationProtection, inst.TerminationProtectionKnown),
		formatProtectionStatus(inst.StopProtection, inst.StopProtectionKnown),
	)

	// Format tags section with a more prominent header
	tagsSection := "\n[::b][yellow]AWS Tags[white][::-]\n"
	if len(inst.Tags) > 0 {
		// Sort tags by key for consistent display
		keys := make([]string, 0, len(inst.Tags))
		for k := range inst.Tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Group tags by category for better organization
		categories := map[string]map[string]string{
			"Resource":  {}, // Resource-related tags (Name, stack info)
			"Business":  {}, // Business context tags (env, project, owner)
			"Technical": {}, // Technical tags (role, version, tier)
			"Other":     {}, // Any other tags that don't fit above categories
		}

		for _, key := range keys {
			value := inst.Tags[key]

			// Categorize tags
			switch strings.ToLower(key) {
			case "name", "aws:cloudformation:stack-name", "aws:cloudformation:logical-id", "aws:autoscaling:groupname":
				categories["Resource"][key] = value
			case "environment", "env", "project", "owner", "team", "cost-center", "application", "app", "service", "product", "costcenter", "business-unit":
				categories["Business"][key] = value
			case "role", "version", "tier", "type", "platform", "auto-delete", "auto-stop", "backup", "cluster", "scheduler":
				categories["Technical"][key] = value
			default:
				categories["Other"][key] = value
			}
		}

		// Display tags by category
		for category, tagMap := range categories {
			if len(tagMap) == 0 {
				continue
			}

			tagsSection += fmt.Sprintf("  [::b][yellow]%s Tags[white][::-]\n", category)

			// Sort keys within category
			catKeys := make([]string, 0, len(tagMap))
			for k := range tagMap {
				catKeys = append(catKeys, k)
			}
			sort.Strings(catKeys)

			// Calculate the longest key for alignment
			longestKey := 0
			for _, key := range catKeys {
				if len(key) > longestKey {
					longestKey = len(key)
				}
			}

			for _, key := range catKeys {
				// Add padding for alignment
				padding := strings.Repeat(" ", longestKey-len(key))
				tagsSection += fmt.Sprintf("    [blue]%s%s:[white] %s\n", key, padding, tagMap[key])
			}

			// Add a blank line between categories
			tagsSection += "\n"
		}
	} else {
		tagsSection += "  No tags found on this instance\n"
	}

	// Combine all sections
	details := baseDetails + tagsSection + "\n[yellow]Press Esc to close[-]"

	detailsText.SetText(details)
	detailsText.SetBorder(true).
		SetTitle(fmt.Sprintf(" Instance: %s ", instance.DisplayName())).
		SetBorderColor(color.AppColors.Border).
		SetTitleColor(color.AppColors.Title)

		// Create a modal that fills most of the screen
		// Make the detail view wider to accommodate tags better
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(detailsText, 80, 1, true).
			AddItem(nil, 0, 1, false), 0, 8, true).
		AddItem(nil, 0, 1, false)

	v.ui.pages.AddPage("modal", flex, true, true)
}

// getStateEmoji returns an emoji representing the instance state
func getStateEmoji(state string) string {
	switch state {
	case "running":
		return "🟢"
	case "stopped":
		return "🔴"
	case "stopping":
		return "🟠"
	case "pending":
		return "🟡"
	case "shutting-down":
		return "💤"
	case "terminated":
		return "⛔"
	case "rebooting":
		return "🔄"
	default:
		return "❓"
	}
}

// getStateColor returns the appropriate color for an instance state
func getStateColor(state string) tcell.Color {
	switch state {
	case "running":
		return color.AppColors.Running
	case "stopped":
		return color.AppColors.Stopped
	case "stopping", "pending", "shutting-down":
		return color.AppColors.Pending
	case "terminated":
		return color.AppColors.Secondary
	case "rebooting":
		return color.AppColors.Pending
	default:
		return color.AppColors.Foreground
	}
}

func formatProtectionStatus(enabled bool, known bool) string {
	if !known {
		return "Unknown"
	}
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func formatProtectionCell(enabled bool, known bool) string {
	if !known {
		return "Unknown"
	}
	if enabled {
		return "On"
	}
	return "Off"
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", h, m)
	} else if d < 30*24*time.Hour {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	} else if d < 365*24*time.Hour {
		months := int(d.Hours() / 24 / 30)
		days := int(d.Hours()/24) % 30
		return fmt.Sprintf("%dM %dd", months, days)
	} else {
		years := int(d.Hours() / 24 / 365)
		months := int(d.Hours()/24/30) % 12
		return fmt.Sprintf("%dy %dM", years, months)
	}
}

// No resize method needed as tview will handle this internally
