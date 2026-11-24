package cmd

import (
	"log"
	"strings"

	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
)

// The ui-* commands are run by Alfred UI script actions (workflow/*.alfredui,
// through workflow/scripts/run.sh). Their stdout must be JSON actions only, so
// they use Run (not RunE: a returned error would make run() call
// wf.FatalError) and never send Script Filter output.

// Component id of the status label every UI has.
const uiStatus = "status"

func emit(a alfredui.Actions) {
	if err := alfredui.Emit(a); err != nil {
		log.Printf("[ui] emit failed: %v", err)
	}
}

// apiMessage shortens go-github's "failed to x: GET https://...: 404 Not
// Found []" to "failed to x: 404 Not Found" for a one-line status label.
func apiMessage(err error) string {
	msg := err.Error()
	if i := strings.Index(msg, "https://"); i >= 0 {
		if j := strings.Index(msg[i:], ": "); j >= 0 {
			prefix := strings.TrimSuffix(strings.TrimRight(msg[:i], " "), "GET")
			msg = strings.TrimRight(prefix, " ") + " " + msg[i+j+2:]
		}
	}
	return strings.TrimSpace(strings.TrimSuffix(msg, " []"))
}
