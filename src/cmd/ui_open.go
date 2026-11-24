package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// envItemURL is set by the repo window's open actions from the selected row.
const envItemURL = "item_url"

// uiOpenCmd opens the selected issue or pull request and closes the window,
// closing the preview sheet first when run from it (finish is ignored while a
// sheet is open). It logs the URL it got, which shows in Alfred's debugger.
var uiOpenCmd = &cobra.Command{
	Use:       "ui-open [sheet]",
	Short:     "open the selected issue or pull request",
	Hidden:    true,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"sheet"},
	Run: func(_ *cobra.Command, args []string) {
		emit(uiOpen(os.Getenv(envItemURL), len(args) > 0 && args[0] == "sheet"))
	},
}

func uiOpen(url string, fromSheet bool) alfredui.Actions {
	log.Printf("[ui] open %q (from sheet: %v)", url, fromSheet)
	a := alfredui.Actions{}
	if fromSheet {
		// finish is ignored while a sheet is open.
		a.Sheet = &alfredui.Sheet{Close: "preview"}
	}
	if !strings.HasPrefix(url, "https://") {
		return status("Nothing selected to open", alfredui.ColourError).WithSheet(a.Sheet)
	}
	if out, err := exec.Command("/usr/bin/open", url).CombinedOutput(); err != nil {
		return status(fmt.Sprintf("open %s: %v: %s", url, err, out), alfredui.ColourError).WithSheet(a.Sheet)
	}
	a.Finish = &alfredui.Finish{}
	return a
}

func init() {
	rootCmd.AddCommand(uiOpenCmd)
}
