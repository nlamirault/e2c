// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"time"
)

// Instance represents an EC2 instance
type Instance struct {
	ID           string            // Instance ID
	Name         string            // Instance name (from Name tag)
	Type         string            // Instance type (e.g., t2.micro)
	State        string            // Current state (running, stopped, etc.)
	Region       string            // AWS region
	LaunchTime   time.Time         // When the instance was launched
	Age          time.Duration     // Age of the instance
	PrivateIP    string            // Private IP address
	PublicIP     string            // Public IP address
	Platform     string            // Platform details (e.g., Linux/UNIX, Windows)
	Architecture string            // Architecture (e.g., x86_64, arm64)
	Tags         map[string]string // AWS tags associated with the instance
	// Protection settings
	TerminationProtection      bool // Whether termination protection is enabled
	StopProtection             bool // Whether stop protection is enabled
	TerminationProtectionKnown bool // Whether termination protection has been fetched
	StopProtectionKnown        bool // Whether stop protection has been fetched
}

// ProtectionStatus represents the protection attributes of an instance.
type ProtectionStatus struct {
	InstanceID            string
	TerminationProtection bool
	StopProtection        bool
}

// GetSSHCommand returns an SSH command for connecting to the instance
func (i *Instance) GetSSHCommand(username string) string {
	ip := i.PublicIP
	if ip == "" {
		ip = i.PrivateIP
	}

	if ip == "" {
		return "No IP address available for SSH connection"
	}

	return fmt.Sprintf("ssh %s@%s", username, ip)
}

// IsRunning returns true if the instance is running
func (i *Instance) IsRunning() bool {
	return i.State == "running"
}

// IsStopped returns true if the instance is stopped
func (i *Instance) IsStopped() bool {
	return i.State == "stopped"
}

// DisplayName returns the name to display (name or ID if name is empty)
func (i *Instance) DisplayName() string {
	if i.Name != "" {
		return i.Name
	}
	return i.ID
}

// StateColor returns the color name to use for the instance state
func (i *Instance) StateColor() string {
	switch i.State {
	case "running":
		return "green"
	case "stopped":
		return "red"
	case "stopping":
		return "yellow"
	case "pending":
		return "yellow"
	case "shutting-down":
		return "yellow"
	case "terminated":
		return "gray"
	default:
		return "white"
	}
}
