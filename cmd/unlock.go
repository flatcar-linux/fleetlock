package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/flatcar-linux/fleetlock/pkg/client"
)

func unlock(group, id, url *string) *cobra.Command {
	return &cobra.Command{
		Use: "unlock-if-held",
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := http.DefaultClient

			c, err := client.New(*url, *group, *id, httpClient)
			if err != nil {
				return fmt.Errorf("building the client: %w", err)
			}

			if err := c.UnlockIfHeld(context.Background()); err != nil {
				return fmt.Errorf("unlocking: %w", err)
			}

			return nil
		},
	}
}
