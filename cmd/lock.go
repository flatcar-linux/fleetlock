package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flatcar-linux/fleetlock/pkg/client"
)

func lock(group, id, url *string) *cobra.Command {
	return &cobra.Command{
		Use:   "recursive-lock",
		Short: "Try to reserve (lock) a slot for rebooting",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkID(id); err != nil {
				return fmt.Errorf("checking ID: %w", err)
			}

			c, err := client.New(&client.Config{
				ID:    *id,
				Group: *group,
				URL:   *url,
			})
			if err != nil {
				return fmt.Errorf("building the client: %w", err)
			}

			if err := c.RecursiveLock(context.Background()); err != nil {
				return fmt.Errorf("locking: %w", err)
			}

			return nil
		},
	}
}
