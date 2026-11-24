package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rwilgaard/alfred-github-search/src/internal/github"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// envAuthToken is set by the Save action in workflow/login.alfredui. The
// token arrives in the environment, never in argv.
const envAuthToken = "auth_token"

// Component ids in workflow/login.alfredui/ui.json.
const (
	uiToken = "token"
	uiSave  = "save"
)

var uiAuthCmd = &cobra.Command{
	Use:    "ui-auth",
	Short:  "store the token from the login form",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		emit(uiAuth(os.Getenv(envAuthToken)))
	},
}

func uiAuth(token string) alfredui.Actions {
	// Undo the busy state ui.json sets before running this.
	idle := alfredui.NewUpdate().
		Enable(true, uiSave).
		Set(uiSave, alfredui.Props{"text": "Save"}).
		Focus(uiToken)

	token = strings.TrimSpace(token)
	if token == "" {
		return idle.Status(uiStatus, "Enter a personal access token.", alfredui.ColourError)
	}

	// MaxAge 0 and a token-keyed cache: the check always reaches GitHub.
	service, err := github.NewAuthenticatedService(token, wf.Cache, 0)
	if err != nil {
		return idle.Status(uiStatus, err.Error(), alfredui.ColourError)
	}
	if err := service.TestAuthentication(); err != nil {
		return idle.Status(uiStatus, fmt.Sprintf("Authentication failed: %s", apiMessage(err)), alfredui.ColourError)
	}
	if err := wf.SetAuth(keychainAccount, token); err != nil {
		return idle.Status(uiStatus, err.Error(), alfredui.ColourError)
	}

	// The finish args go to the notification after the Alfred UI object.
	return alfredui.Actions{Finish: &alfredui.Finish{Args: "Successfully authenticated"}}
}

func init() {
	rootCmd.AddCommand(uiAuthCmd)
}
