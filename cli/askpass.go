package cli

import (
	"fmt"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	"github.com/coder/coder/cli/gitaskpass"
	"github.com/coder/coder/codersdk"
)

func askpass() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "askpass",
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()

			ctx, stop := signal.NotifyContext(ctx, interruptSignals...)
			defer stop()

			defer func() {
				if ctx.Err() != nil {
					err = ctx.Err()
				}
			}()

			user, host, err := gitaskpass.Parse(args[0])
			if err != nil {
				return xerrors.Errorf("parse host: %w", err)
			}

			client, err := createAgentClient(cmd)
			if err != nil {
				return xerrors.Errorf("create agent client: %w", err)
			}

			apResp, err := client.WorkspaceAgentRequestGitAuth(ctx, codersdk.WorkspaceAgentGitAuthRequest{
				User: user,
				URL:  host,
			})
			if err != nil {
				return xerrors.Errorf("workspace agent askpass: %w", err)
			}

			if askpassAuthPending(apResp) {
				fmt.Fprintf(cmd.ErrOrStderr(), "Please visit %s to authenticate.\n", apResp.AuthURL)

				for askpassAuthPending(apResp) {
					time.Sleep(time.Second) // TODO(mafredri): Replace with websockets or long polling...
					apResp, err = client.WorkspaceAgentGitAuthRequest(ctx, apResp.RequestID)
					if err != nil {
						return xerrors.Errorf("workspace agent askpass request: %w", err)
					}
				}

				fmt.Fprintf(cmd.ErrOrStderr(), "Authentication complete!\n")
			}

			if user == "" {
				fmt.Println(apResp.User)
			} else {
				fmt.Println(apResp.AccessToken)
			}

			return nil
		},
	}

	return cmd
}

func askpassAuthPending(resp codersdk.WorkspaceAgentGitAuthResponse) bool {
	return resp.AccessToken == ""
}
