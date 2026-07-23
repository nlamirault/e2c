// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/nlamirault/e2c/internal/aws"
	"github.com/nlamirault/e2c/internal/color"
	"github.com/nlamirault/e2c/internal/config"
	"github.com/nlamirault/e2c/internal/model"
)

// UI manages the terminal UI for e2c
type UI struct {
	app           *tview.Application
	pages         *tview.Pages
	instancesView *InstancesView
	overviewPanel *OverviewPanel
	statusBar     *StatusBar
	helpView      *HelpView
	log           *slog.Logger
	ec2Client     *aws.EC2Client
	config        *config.Config
	ctx           context.Context
	cancel        context.CancelFunc
	refreshTicker *time.Ticker
	refreshMutex  sync.Mutex
	filter        string
}

// NewUI creates a new UI instance
func NewUI(log *slog.Logger, ec2Client *aws.EC2Client, cfg *config.Config) *UI {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize colors
	color.InitializeColors()

	ui := &UI{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		log:       log,
		ec2Client: ec2Client,
		config:    cfg,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize components
	ui.instancesView = NewInstancesView(ui)
	ui.overviewPanel = NewOverviewPanel(ui)
	ui.statusBar = NewStatusBar(ui)
	ui.helpView = NewHelpView(cfg.UI.ExpertMode)

	// Set initial region in status bar
	ui.statusBar.SetRegion(ec2Client.GetRegion())

	// Set up the main layout
	ui.setupLayout()

	// Set up key bindings
	ui.setupKeyBindings()

	return ui
}

// Start starts the UI
func (ui *UI) Start() error {
	// Set initial UI components
	ui.statusBar.SetMode("normal")

	// Start refresh ticker
	ui.startRefreshTicker()

	// Initial data load
	ui.RefreshInstances()

	// Run the application
	if err := ui.app.Run(); err != nil {
		return fmt.Errorf("error running application: %w", err)
	}

	return nil
}

// We'll simplify our approach by letting the tview layout system handle sizing

// Stop stops the UI
func (ui *UI) Stop() {
	ui.cancel()
	if ui.refreshTicker != nil {
		ui.refreshTicker.Stop()
	}
	ui.app.Stop()
}

// setupLayout sets up the main layout of the application
func (ui *UI) setupLayout() {
	// Create main layout
	grid := tview.NewGrid().
		SetRows(5, 0, 1, 1). // Overview panel, main content, status bar, help
		SetColumns(0).       // Full width
		SetBorders(false)

	// Set instance table title with theme colors
	// Set instance table title with theme colors
	ui.instancesView.table.
		SetTitle("Instances").
		SetBorder(true).
		SetBorderColor(color.AppColors.Border)

	// Add components to the grid with proper proportions
	grid.AddItem(ui.overviewPanel.view, 0, 0, 1, 1, 0, 0, false).
		AddItem(ui.instancesView.table, 1, 0, 1, 1, 0, 0, true).
		AddItem(ui.statusBar.view, 2, 0, 1, 1, 0, 0, false).
		AddItem(ui.helpView.view, 3, 0, 1, 1, 0, 0, false)

	// Add main page
	ui.pages.AddPage("main", grid, true, true)

	// Set the root of the application
	ui.app.SetRoot(ui.pages, true)
}

// setupKeyBindings sets up the global key bindings
func (ui *UI) setupKeyBindings() {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Global key bindings
		switch event.Key() {
		case tcell.KeyEscape:
			// Go back to main page if on a modal
			if ui.pages.HasPage("modal") {
				ui.pages.RemovePage("modal")
				return nil
			}
		}

		// Process based on current page
		name, _ := ui.pages.GetFrontPage()
		switch {
		case ui.pages.HasPage("main") && name == "main":
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q':
					ui.Stop()
					return nil
				case 'r':
					ui.RefreshInstances()
					return nil
				case 'f':
					ui.ShowFilterDialog()
					return nil
				case '?':
					ui.ShowHelpDialog()
					return nil
				case 's':
					ui.handleStartInstance()
					return nil
				case 'p':
					ui.handleStopInstance()
					return nil
				case 'b':
					ui.handleRebootInstance()
					return nil
				case 't':
					ui.handleTerminateInstance()
					return nil
				case 'c':
					ui.handleConnectInstance()
					return nil
				case 'l':
					ui.handleViewLogs()
					return nil
				case 'x':
					if ui.config.UI.ExpertMode {
						ui.handleToggleTerminationProtection()
						return nil
					}
				case 'n':
					if ui.config.UI.ExpertMode {
						ui.handleToggleStopProtection()
						return nil
					}
				}
			}
		}
		return event
	})
}

// RefreshInstances refreshes the instances list
func (ui *UI) RefreshInstances() {
	_ = ui.instancesView.GetSelectedInstance()
	ui.refreshMutex.Lock()
	defer ui.refreshMutex.Unlock()

	ui.statusBar.SetStatus("Refreshing instances...")

	go func() {
		instances, err := ui.ec2Client.ListInstances(ui.ctx, ui.config.UI.ExpertMode)
		if err != nil {
			ui.log.Error("Failed to list instances", "error", err)
			ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
			return
		}

		// Count running and stopped instances
		running := 0
		stopped := 0
		for _, instance := range instances {
			if instance.IsRunning() {
				running++
			} else if instance.IsStopped() {
				stopped++
			}
		}

		// Apply filter if present
		filteredInstances := ui.applyFilter(instances)

		// Update UI with instances
		ui.app.QueueUpdateDraw(func() {
			ui.instancesView.UpdateInstances(filteredInstances)
			ui.overviewPanel.Update(len(instances), running, stopped, ui.ec2Client.GetRegion())
			ui.statusBar.SetRegion(ui.ec2Client.GetRegion())
			ui.statusBar.SetStatus(fmt.Sprintf("Found %d instances", len(filteredInstances)))
		})

		if ui.config.UI.ExpertMode {
			ui.fetchProtectionsInBackground(filteredInstances)
		}
	}()
}

func (ui *UI) fetchProtectionsInBackground(instances []model.Instance) {
	idsToFetch := make([]string, 0, len(instances))
	for _, inst := range instances {
		if _, _, ok := ui.ec2Client.GetCachedProtectionStatus(inst.ID); ok {
			continue
		}
		idsToFetch = append(idsToFetch, inst.ID)
	}

	if len(idsToFetch) == 0 {
		return
	}

	go func() {
		for status := range ui.ec2Client.FetchProtectionStatuses(ui.ctx, idsToFetch, 5) {
			ui.app.QueueUpdateDraw(func() {
				ui.instancesView.UpdateProtection(status.InstanceID, status.TerminationProtection, status.StopProtection)
			})
		}
	}()
}

// startRefreshTicker starts a ticker to refresh instances periodically
func (ui *UI) startRefreshTicker() {
	interval := ui.config.AWS.RefreshInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ui.refreshTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ui.refreshTicker.C:
				ui.RefreshInstances()
			case <-ui.ctx.Done():
				return
			}
		}
	}()
}

// applyFilter applies the current filter to instances
func (ui *UI) applyFilter(instances []model.Instance) []model.Instance {
	if ui.filter == "" {
		return instances
	}

	filtered := make([]model.Instance, 0)
	for _, instance := range instances {
		if ui.matchesFilter(instance) {
			filtered = append(filtered, instance)
		}
	}

	return filtered
}

// matchesFilter checks if an instance matches the current filter
func (ui *UI) matchesFilter(instance model.Instance) bool {
	filter := ui.filter
	if filter == "" {
		return true
	}

	// Match against various fields
	return containsIgnoreCase(instance.ID, filter) ||
		containsIgnoreCase(instance.Name, filter) ||
		containsIgnoreCase(instance.Type, filter) ||
		containsIgnoreCase(instance.State, filter) ||
		containsIgnoreCase(instance.PrivateIP, filter) ||
		containsIgnoreCase(instance.PublicIP, filter)
}

// SetFilter sets the instance filter
func (ui *UI) SetFilter(filter string) {
	ui.filter = filter
	ui.RefreshInstances()
}

// ShowFilterDialog displays the filter dialog
func (ui *UI) ShowFilterDialog() {
	// Set UI mode to filtering
	ui.statusBar.SetMode("filtering")

	form := tview.NewForm()
	form.AddInputField("Filter:", ui.filter, 30, nil, nil)
	form.AddButton("Apply", func() {
		filter := form.GetFormItem(0).(*tview.InputField).GetText()
		ui.SetFilter(filter)
		ui.statusBar.SetMode("normal")
		ui.pages.RemovePage("modal")
	})
	form.AddButton("Clear", func() {
		ui.SetFilter("")
		ui.statusBar.SetMode("normal")
		ui.pages.RemovePage("modal")
	})
	form.AddButton("Cancel", func() {
		ui.statusBar.SetMode("normal")
		ui.pages.RemovePage("modal")
	})

	form.SetBorder(true).SetTitle("Filter Instances")
	form.SetCancelFunc(func() {
		ui.statusBar.SetMode("normal")
		ui.pages.RemovePage("modal")
	})

	// Use a reasonable fixed width for the form
	formWidth := 40

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, formWidth, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("modal", flex, true, true)
}

// ShowHelpDialog displays the help dialog

// GetColors returns the application colors
func (ui *UI) GetColors() color.Colors {
	return color.AppColors
}

func (ui *UI) ShowHelpDialog() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	expertShortcuts := ""
	if ui.config.UI.ExpertMode {
		expertShortcuts = "  [green]x[white]      Toggle termination protection[-]\n" +
			"  [green]n[white]      Toggle stop protection[-]\n"
	}

	helpText.SetText(fmt.Sprintf(`
[::b]e2c - AWS EC2 Terminal UI Manager[::-]

[yellow]Keyboard Shortcuts:[-]
  [green]?[white]      Help (this screen)[-]
  [green]q[white]      Quit[-]
  [green]r[white]      Refresh instances[-]
  [green]f[white]      Filter instances[-]
  [green]s[white]      Start selected instance[-]
  [green]p[white]      Stop selected instance[-]
  [green]b[white]      Reboot selected instance[-]
  [green]t[white]      Terminate selected instance[-]
  [green]c[white]      Connect to selected instance via SSH[-]
  [green]l[white]      View instance logs/console output[-]
%s  [green]Esc[white]    Close dialogs[-]

[yellow]Press Esc to close this help[-]
`, expertShortcuts))

	helpText.SetBorder(true).SetTitle("Help")

	// Use fixed dimensions that work well in most terminals
	// modalWidth := 70

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(helpText, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	// The flex layout will handle positioning

	ui.pages.AddPage("modal", flex, true, true)
}

// ShowConfirmDialog shows a confirmation dialog
func (ui *UI) ShowConfirmDialog(title, message string, onConfirm func()) {
	// Use a reasonable fixed width
	modalWidth := 60

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ui.pages.RemovePage("modal")
			if buttonLabel == "Yes" {
				onConfirm()
			}
		})

	modal.SetBorder(true).SetTitle(title).SetBorderColor(tcell.ColorBlue)

	// Create a flex container to position the modal
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(modal, modalWidth, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("modal", flex, true, true)
}

// ShowInfoDialog shows an information dialog
func (ui *UI) ShowInfoDialog(title, message string) {
	// Use a reasonable fixed width
	modalWidth := 60

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ui.pages.RemovePage("modal")
		})

	modal.SetBorder(true).SetTitle(title).SetBorderColor(tcell.ColorBlue)

	// Create a flex container to position the modal
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(modal, modalWidth, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("modal", flex, true, true)
}

// handleStartInstance handles starting the selected instance
func (ui *UI) handleStartInstance() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	if selectedInstance.IsRunning() {
		ui.statusBar.SetError("Instance is already running")
		return
	}

	ui.ShowConfirmDialog(
		"Start Instance",
		fmt.Sprintf("Are you sure you want to start instance %s?", selectedInstance.DisplayName()),
		func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Starting instance %s...", selectedInstance.ID))

			go func() {
				err := ui.ec2Client.StartInstance(ui.ctx, selectedInstance.ID)
				if err != nil {
					ui.app.QueueUpdateDraw(func() {
						ui.log.Error("Failed to start instance", "error", err)
						ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
					})
					return
				}

				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetStatus(fmt.Sprintf("Started instance %s", selectedInstance.ID))
					ui.RefreshInstances()
				})
			}()
		},
	)
}

// handleStopInstance handles stopping the selected instance
func (ui *UI) handleStopInstance() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	if selectedInstance.IsStopped() {
		ui.statusBar.SetError("Instance is already stopped")
		return
	}

	ui.ShowConfirmDialog(
		"Stop Instance",
		fmt.Sprintf("Are you sure you want to stop instance %s?", selectedInstance.DisplayName()),
		func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Stopping instance %s...", selectedInstance.ID))

			go func() {
				err := ui.ec2Client.StopInstance(ui.ctx, selectedInstance.ID)
				if err != nil {
					ui.app.QueueUpdateDraw(func() {
						ui.log.Error("Failed to stop instance", "error", err)
						ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
					})
					return
				}

				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetStatus(fmt.Sprintf("Stopped instance %s", selectedInstance.ID))
					ui.RefreshInstances()
				})
			}()
		},
	)
}

// handleRebootInstance handles rebooting the selected instance
func (ui *UI) handleRebootInstance() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	if !selectedInstance.IsRunning() {
		ui.statusBar.SetError("Instance must be running to reboot")
		return
	}

	ui.ShowConfirmDialog(
		"Reboot Instance",
		fmt.Sprintf("Are you sure you want to reboot instance %s?", selectedInstance.DisplayName()),
		func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Rebooting instance %s...", selectedInstance.ID))

			go func() {
				err := ui.ec2Client.RebootInstance(ui.ctx, selectedInstance.ID)
				if err != nil {
					ui.app.QueueUpdateDraw(func() {
						ui.log.Error("Failed to reboot instance", "error", err)
						ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
					})
					return
				}

				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetStatus(fmt.Sprintf("Rebooted instance %s", selectedInstance.ID))
					ui.RefreshInstances()
				})
			}()
		},
	)
}

// handleTerminateInstance handles terminating the selected instance
func (ui *UI) handleTerminateInstance() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	ui.ShowConfirmDialog(
		"Terminate Instance",
		fmt.Sprintf("Are you sure you want to TERMINATE instance %s? This action cannot be undone!", selectedInstance.DisplayName()),
		func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Terminating instance %s...", selectedInstance.ID))

			go func() {
				err := ui.ec2Client.TerminateInstance(ui.ctx, selectedInstance.ID)
				if err != nil {
					ui.app.QueueUpdateDraw(func() {
						ui.log.Error("Failed to terminate instance", "error", err)
						ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
					})
					return
				}

				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetStatus(fmt.Sprintf("Terminated instance %s", selectedInstance.ID))
					ui.RefreshInstances()
				})
			}()
		},
	)
}

// handleToggleTerminationProtection toggles termination protection on the selected instance
func (ui *UI) handleToggleTerminationProtection() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	go func() {
		termState := selectedInstance.TerminationProtection
		knownTerm := selectedInstance.TerminationProtectionKnown
		knownStop := selectedInstance.StopProtectionKnown

		if !knownTerm || !knownStop {
			refreshedTerm, _, err := ui.ec2Client.RefreshProtectionStatus(ui.ctx, selectedInstance.ID)
			if err != nil {
				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetError(fmt.Sprintf("Failed to reload protections: %v", err))
				})
				return
			}

			termState = refreshedTerm
			ui.app.QueueUpdateDraw(func() {
				ui.instancesView.UpdateProtection(selectedInstance.ID, refreshedTerm, selectedInstance.StopProtection)
			})
		}

		targetState := !termState
		action := "Disabling"
		if targetState {
			action = "Enabling"
		}

		ui.app.QueueUpdateDraw(func() {
			ui.statusBar.SetStatus(fmt.Sprintf("%s termination protection for %s...", action, selectedInstance.ID))
		})

		err := ui.ec2Client.SetTerminationProtection(ui.ctx, selectedInstance.ID, targetState)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.log.Error("Failed to update termination protection", "error", err)
				ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
			})
			return
		}

		term, stop, err := ui.ec2Client.RefreshProtectionStatus(ui.ctx, selectedInstance.ID)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.statusBar.SetError(fmt.Sprintf("Failed to reload protections: %v", err))
			})
			return
		}

		ui.app.QueueUpdateDraw(func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Termination protection %s for %s", protectionStatusText(targetState), selectedInstance.ID))
			ui.instancesView.UpdateProtection(selectedInstance.ID, term, stop)
		})
	}()
}

// handleToggleStopProtection toggles stop protection on the selected instance
func (ui *UI) handleToggleStopProtection() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	go func() {
		stopState := selectedInstance.StopProtection
		knownTerm := selectedInstance.TerminationProtectionKnown
		knownStop := selectedInstance.StopProtectionKnown

		if !knownTerm || !knownStop {
			_, refreshedStop, err := ui.ec2Client.RefreshProtectionStatus(ui.ctx, selectedInstance.ID)
			if err != nil {
				ui.app.QueueUpdateDraw(func() {
					ui.statusBar.SetError(fmt.Sprintf("Failed to reload protections: %v", err))
				})
				return
			}

			stopState = refreshedStop
			ui.app.QueueUpdateDraw(func() {
				ui.instancesView.UpdateProtection(selectedInstance.ID, selectedInstance.TerminationProtection, refreshedStop)
			})
		}

		targetState := !stopState
		action := "Disabling"
		if targetState {
			action = "Enabling"
		}

		ui.app.QueueUpdateDraw(func() {
			ui.statusBar.SetStatus(fmt.Sprintf("%s stop protection for %s...", action, selectedInstance.ID))
		})

		err := ui.ec2Client.SetStopProtection(ui.ctx, selectedInstance.ID, targetState)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.log.Error("Failed to update stop protection", "error", err)
				ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
			})
			return
		}

		term, stop, err := ui.ec2Client.RefreshProtectionStatus(ui.ctx, selectedInstance.ID)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.statusBar.SetError(fmt.Sprintf("Failed to reload protections: %v", err))
			})
			return
		}

		ui.app.QueueUpdateDraw(func() {
			ui.statusBar.SetStatus(fmt.Sprintf("Stop protection %s for %s", protectionStatusText(targetState), selectedInstance.ID))
			ui.instancesView.UpdateProtection(selectedInstance.ID, term, stop)
		})
	}()
}

// handleConnectInstance handles connecting to the selected instance
func (ui *UI) handleConnectInstance() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	if !selectedInstance.IsRunning() {
		ui.statusBar.SetError("Instance must be running to connect")
		return
	}

	// Default username based on platform
	defaultUser := "ec2-user"
	if selectedInstance.Platform != "" {
		if containsIgnoreCase(selectedInstance.Platform, "ubuntu") {
			defaultUser = "ubuntu"
		} else if containsIgnoreCase(selectedInstance.Platform, "debian") {
			defaultUser = "admin"
		} else if containsIgnoreCase(selectedInstance.Platform, "windows") {
			defaultUser = "Administrator"
		}
	}

	form := tview.NewForm()
	form.AddInputField("Username:", defaultUser, 20, nil, nil)
	form.AddButton("Connect", func() {
		username := form.GetFormItem(0).(*tview.InputField).GetText()
		sshCommand := selectedInstance.GetSSHCommand(username)

		ui.ShowInfoDialog("SSH Command", sshCommand)
	})
	form.AddButton("Cancel", func() {
		ui.pages.RemovePage("modal")
	})

	form.SetBorder(true).SetTitle("SSH Connection")

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 40, 1, true).
			AddItem(nil, 0, 1, false), 0, 8, true).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("modal", flex, true, true)
}

// handleViewLogs handles viewing the console output of the selected instance
func (ui *UI) handleViewLogs() {
	selectedInstance := ui.instancesView.GetSelectedInstance()
	if selectedInstance == nil {
		ui.statusBar.SetError("No instance selected")
		return
	}

	ui.statusBar.SetStatus(fmt.Sprintf("Fetching console output for instance %s...", selectedInstance.ID))

	go func() {
		output, err := ui.ec2Client.GetInstanceConsoleOutput(ui.ctx, selectedInstance.ID)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.log.Error("Failed to get console output", "error", err)
				ui.statusBar.SetError(fmt.Sprintf("Error: %v", err))
			})
			return
		}

		ui.app.QueueUpdateDraw(func() {
			ui.statusBar.SetStatus("Showing console output")

			textView := tview.NewTextView().
				SetDynamicColors(true).
				SetScrollable(true).
				SetText(output)

			textView.SetBorder(true).SetTitle(fmt.Sprintf("Console Output: %s", selectedInstance.DisplayName()))

			// Center the text view
			flex := tview.NewFlex().
				AddItem(nil, 0, 1, false).
				AddItem(tview.NewFlex().
					AddItem(nil, 0, 1, false).
					AddItem(textView, 80, 1, true).
					AddItem(nil, 0, 1, false), 0, 8, true).
				AddItem(nil, 0, 1, false)

			flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					ui.pages.RemovePage("modal")
					return nil
				}
				return event
			})

			ui.pages.AddPage("modal", flex, true, true)
		})
	}()
}

// containsIgnoreCase checks if a string contains another string, ignoring case
func containsIgnoreCase(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	return fmt.Sprintf("%s", s) != "" &&
		containsRune(fmt.Sprintf("%s", s), fmt.Sprintf("%s", substr))
}

func protectionStatusText(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// containsRune is a simple case-insensitive substring check
func containsRune(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}

	s = toLower(s)
	substr = toLower(substr)

	return indexString(s, substr) >= 0
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + ('a' - 'A'))
		} else {
			result += string(r)
		}
	}
	return result
}

// indexString finds the index of substr in s
func indexString(s, substr string) int {
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
