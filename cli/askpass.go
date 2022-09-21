package cli

import (
	"fmt"

	"github.com/coder/coder/cli/gitaskpass"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

// askpass is used to internally swap
func askpass() *cobra.Command {
	return &cobra.Command{
		Use:  "askpass",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt, host, err := gitaskpass.Parse(args[0])
			if err != nil {
				return xerrors.Errorf("parse host: %w", err)
			}
			if prompt != "Username" {
				return nil
			}

			fmt.Printf("Host: %s\n", host)
			// We should request coderd for the authentication token, and
			// on a specific status code we get a different response.

			return xerrors.New("we here")
		},
	}
}
