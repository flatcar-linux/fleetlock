// Package cmd implements fleetlockctl CLI.
package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

// Command returns CLI command to be executed.
func Command() *cobra.Command {
	cli := &cobra.Command{Use: "fleetlockctl"}

	var group, id, url string

	cli.PersistentFlags().StringVarP(&group, "group", "g", "default", "FleetLock group")
	cli.PersistentFlags().StringVarP(&id, "id", "i", "", "FleetLock instance ID (e.g. content of /etc/machine-id file)")
	cli.PersistentFlags().StringVarP(&url, "url", "u", "", "FleetLock endpoint URL")

	cli.AddCommand(lock(&group, &id, &url))
	cli.AddCommand(unlock(&group, &id, &url))

	return cli
}

// machineID is a helper to return unique ID
// of the machine.
func machineID() (string, error) {
	id, err := ioutil.ReadFile("/etc/machine-id")
	if err != nil {
		return "", fmt.Errorf("reading machine ID from file: %w", err)
	}

	return string(id), nil
}

// checkID asserts that the ID is not nil, if it's the case
// it uses `machineID` to generate a default one.
func checkID(id *string) error {
	// the ID is set and it's not empty.
	if id != nil && *id != "" {
		return nil
	}

	i, err := machineID()
	if err != nil {
		return fmt.Errorf("getting default machine ID: %w", err)
	}

	*id = i

	return nil
}
