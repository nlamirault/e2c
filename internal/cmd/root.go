package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nlamirault/e2c/internal/aws"
	"github.com/nlamirault/e2c/internal/ui"
	"github.com/nlamirault/e2c/internal/version"
	"github.com/nlamirault/e2c/pkg/config"
)

// NewRootCommand creates the root command for e2c
func NewRootCommand(log *logrus.Logger) *cobra.Command {
	var (
		cfgFile string
		profile string
		region  string
	)

	cmd := &cobra.Command{
		Use:   "e2c",
		Short: "AWS EC2 Terminal UI Manager",
		Long: `e2c is a terminal-based UI application for managing AWS EC2 instances,
inspired by k9s for Kubernetes and e1s for ECS.
		
It provides a simple, intuitive interface for managing EC2 instances
across multiple regions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(log)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override with CLI flags
			cfg.Override(profile, region)

			// Create AWS EC2 client
			ec2Client, err := aws.NewEC2Client(log, cfg.AWS.DefaultRegion, cfg.AWS.Profile)
			if err != nil {
				return fmt.Errorf("failed to create EC2 client: %w", err)
			}

			// Create and start UI
			app := ui.NewUI(log, ec2Client, cfg)
			if err := app.Start(); err != nil {
				return fmt.Errorf("UI error: %w", err)
			}

			return nil
		},
		Version: version.GetVersion(),
	}

	// Add flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/e2c/config.yaml)")
	cmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile to use")
	cmd.PersistentFlags().StringVar(&region, "region", "", "AWS region to use")

	// Add version command
	cmd.AddCommand(newVersionCommand())

	return cmd
}

// newVersionCommand creates a version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("e2c version %s\n", version.GetVersion())
		},
	}
}

// Execute executes the root command
func Execute() {
	log := logrus.New()
	if err := NewRootCommand(log).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}