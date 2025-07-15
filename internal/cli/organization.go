package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/internal/cli/display"
	"github.com/flightctl/flightctl/internal/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type OrganizationOptions struct {
	GlobalOptions
	Output string
}

func DefaultOrganizationOptions() *OrganizationOptions {
	return &OrganizationOptions{
		GlobalOptions: DefaultGlobalOptions(),
		Output:        "",
	}
}

// NewCmdOrganization creates the main 'org' command
func NewCmdOrganization() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org",
		Aliases: []string{"organization"},
		Short:   "Manage organization context",
		Long:    "Commands for managing organization context. Use 'get orgs' to list available organizations.",
	}

	cmd.AddCommand(NewCmdOrgSelect())
	cmd.AddCommand(NewCmdOrgCurrent())

	return cmd
}

// NewCmdOrgSelect creates the 'org select' command
func NewCmdOrgSelect() *cobra.Command {
	o := DefaultOrganizationOptions()
	cmd := &cobra.Command{
		Use:   "select <org_uuid>",
		Short: "Select an organization and write it to configuration",
		Long:  "Selects an organization and writes it to configuration, allowing the org_id to be set for subsequent requests.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			ctx, cancel := o.WithTimeout(cmd.Context())
			defer cancel()
			return o.RunSelect(ctx, args[0])
		},
		SilenceUsage: true,
	}
	o.Bind(cmd.Flags())
	return cmd
}

// NewCmdOrgCurrent creates the 'org current' command
func NewCmdOrgCurrent() *cobra.Command {
	o := DefaultOrganizationOptions()
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the current organization in context",
		Long:  "Utility for showing the current organization in context.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			ctx, cancel := o.WithTimeout(cmd.Context())
			defer cancel()
			return o.RunCurrent(ctx)
		},
		SilenceUsage: true,
	}
	o.Bind(cmd.Flags())
	return cmd
}

func (o *OrganizationOptions) Bind(fs *pflag.FlagSet) {
	o.GlobalOptions.Bind(fs)
	fs.StringVarP(&o.Output, "output", "o", o.Output, fmt.Sprintf("Output format. One of: (%s).", strings.Join([]string{string(display.JSONFormat), string(display.YAMLFormat), string(display.NameFormat), string(display.WideFormat)}, ", ")))
}

func (o *OrganizationOptions) Complete(cmd *cobra.Command, args []string) error {
	return o.GlobalOptions.Complete(cmd, args)
}

func (o *OrganizationOptions) Validate(args []string) error {
	return o.GlobalOptions.Validate(args)
}

// RunSelect implements the 'org select' command
func (o *OrganizationOptions) RunSelect(ctx context.Context, orgID string) error {
	// Validate UUID format
	if _, err := uuid.Parse(orgID); err != nil {
		return fmt.Errorf("invalid organization UUID format: %s", orgID)
	}

	// Create client to verify organization exists
	c, err := client.NewFromConfigFile(o.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// List organizations to verify the provided ID exists
	resp, err := c.ListUserOrganizationsWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("verifying organization: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to verify organization: %s", resp.Status())
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected empty response")
	}

	// Check if the organization exists
	var selectedOrg *api.Organization
	for _, org := range resp.JSON200.Items {
		if org.Id.String() == orgID {
			selectedOrg = &org
			break
		}
	}

	if selectedOrg == nil {
		return fmt.Errorf("organization %s not found or not accessible", orgID)
	}

	// Load and update the configuration
	config, err := client.ParseConfigFile(o.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Set the organization in the config
	config.SetCurrentOrganization(selectedOrg.Id.String())

	// Persist the updated configuration
	if err := config.Persist(o.ConfigFilePath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	fmt.Printf("Selected organization: %s (%s)\n", selectedOrg.DisplayName, selectedOrg.Id.String())
	return nil
}

// RunCurrent implements the 'org current' command
func (o *OrganizationOptions) RunCurrent(ctx context.Context) error {
	// Load the configuration
	config, err := client.ParseConfigFile(o.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Check if an organization is currently selected
	if !config.HasOrganization() {
		fmt.Println("No organization currently selected")
		return nil
	}

	currentOrgID := config.GetCurrentOrganization()

	// If output format is specified, handle different formats
	if o.Output != "" {
		// For non-table formats, create a simple organization object
		orgInfo := map[string]string{
			"id": currentOrgID,
		}

		switch display.OutputFormat(o.Output) {
		case display.JSONFormat:
			formatter := display.NewFormatter(display.JSONFormat)
			return formatter.Format(orgInfo, display.FormatOptions{Writer: os.Stdout})
		case display.YAMLFormat:
			formatter := display.NewFormatter(display.YAMLFormat)
			return formatter.Format(orgInfo, display.FormatOptions{Writer: os.Stdout})
		case display.NameFormat:
			fmt.Println(currentOrgID)
			return nil
		}
	}

	// For table format or default, try to get organization details
	c, err := client.NewFromConfigFile(o.ConfigFilePath)
	if err != nil {
		// If we can't create a client, just show the ID
		fmt.Printf("Current organization: %s\n", currentOrgID)
		return nil
	}

	// List organizations to get the display name
	resp, err := c.ListUserOrganizationsWithResponse(ctx)
	if err != nil {
		// If we can't get the list, just show the ID
		fmt.Printf("Current organization: %s\n", currentOrgID)
		return nil
	}

	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		// If we can't get the list, just show the ID
		fmt.Printf("Current organization: %s\n", currentOrgID)
		return nil
	}

	// Find the organization details
	for _, org := range resp.JSON200.Items {
		if org.Id.String() == currentOrgID {
			fmt.Printf("Current organization: %s (%s)\n", org.DisplayName, org.Id.String())
			return nil
		}
	}

	// Organization not found in the list (maybe access was revoked)
	fmt.Printf("Current organization: %s (details unavailable)\n", currentOrgID)
	return nil
}
