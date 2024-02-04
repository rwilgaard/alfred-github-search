package alfred

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"

    aw "github.com/deanishe/awgo"
)

type MagicAuth struct {
    Workflow *aw.Workflow
    Account  string
}

func (a MagicAuth) Keyword() string     { return "clearauth" }
func (a MagicAuth) Description() string { return "Clear credentials." }
func (a MagicAuth) RunText() string     { return "Credentials cleared!" }
func (a MagicAuth) Run() error          { return clearAuth(a.Workflow, a.Account) }

func clearAuth(wf *aw.Workflow, keychainAccount string) error {
    if err := wf.Keychain.Delete(keychainAccount); err != nil {
        return err
    }
    return nil
}

func InitWorkflow(wf *aw.Workflow, cfg interface{}) error {
    updateJobName := "checkForUpdates"

    if err := wf.Config.To(cfg); err != nil {
        return err
    }

    if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
        log.Println("Running update check in background...")
        cmd := exec.Command(os.Args[0], "update")
        if err := wf.RunInBackground(updateJobName, cmd); err != nil {
            return fmt.Errorf("Error starting update check: %s", err)
        }
    }

    if wf.UpdateAvailable() {
        wf.Configure(aw.SuppressUIDs(true))
        wf.NewItem("Update Available!").
            Subtitle("Press ⏎ to install").
            Autocomplete("workflow:update").
            Valid(false).
            Icon(aw.IconInfo)
    }

    return nil
}

func HandleFeedback(wf *aw.Workflow) {
    if wf.IsEmpty() {
        wf.NewItem("No results found...").
            Subtitle("Try a different query?").
            Icon(aw.IconInfo)
    }
    wf.SendFeedback()
}

func LoadCache(wf *aw.Workflow, name string, out interface{}) error {
    if wf.Cache.Exists(name) {
        if err := wf.Cache.LoadJSON(name, out); err != nil {
            return err
        }
    }
    return nil
}

func RefreshCache(wf *aw.Workflow, name string, maxAge time.Duration) error {
    if wf.Cache.Expired(name, maxAge) {
        wf.Rerun(2)
        if !wf.IsRunning("cache") {
            cmd := exec.Command(os.Args[0], "cache")
            if err := wf.RunInBackground("cache", cmd); err != nil {
                return err
            }
        } else {
            log.Printf("cache job already running.")
        }

        var cache []interface{}
        err := LoadCache(wf, name, cache)
        if err != nil {
            return err
        }

        if len(cache) == 0 {
            wf.NewItem("Refreshing repository cache…").
                Icon(aw.IconInfo)
            HandleFeedback(wf)
        }
    }
    return nil
}

func HandleAuthentication(wf *aw.Workflow, keychainAccount string) {
    _, err := wf.Keychain.Get(keychainAccount)
    if err != nil {
        wf.NewItem("You're not logged in.").
            Subtitle("Press ⏎ to authenticate").
            Icon(aw.IconInfo).
            Arg("auth").
            Valid(true)
        HandleFeedback(wf)
    }
}
