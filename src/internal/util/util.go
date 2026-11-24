package util

import (
	"os"
	"strconv"
	"strings"

	"github.com/rwilgaard/go-alfredutils/v2/alfred"
)

const (
	iconVariantLight = "light"
	iconVariantDark  = "dark"

	// Midpoint of the 0-255 luminance range.
	lightThemeLuminance = 127.5
)

// Icons that ship with the workflow, relative to workflow/. Each exists in a
// dark-theme and a light-theme variant; the variant follows the Alfred theme
// the process was started under.
var (
	variant = iconVariant()

	IconBack        = icon("back")
	IconBrowser     = icon("browser")
	IconCopy        = icon("copy")
	IconIssueOpen   = icon("issue-open")
	IconPROpen      = icon("pr-open")
	IconRefresh     = icon("refresh")
	IconRepoPrivate = icon("repo-private")
	IconRepoPublic  = icon("repo-public")
)

func icon(name string) alfred.Icon {
	return alfred.Icon{Path: "icons/" + variant + "/" + name + ".png"}
}

func iconVariant() string {
	lum, ok := backgroundLuminance(os.Getenv("alfred_theme_background"))
	if ok && lum > lightThemeLuminance {
		return iconVariantLight
	}
	return iconVariantDark
}

// backgroundLuminance parses an "rgba(r,g,b,a)" string into a perceived
// luminance in the range 0-255.
func backgroundLuminance(bg string) (float64, bool) {
	open := strings.IndexByte(bg, '(')
	closed := strings.LastIndexByte(bg, ')')
	if open < 0 || closed < open {
		return 0, false
	}

	parts := strings.Split(bg[open+1:closed], ",")
	if len(parts) < 3 {
		return 0, false
	}

	var rgb [3]float64
	for i := range rgb {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return 0, false
		}
		rgb[i] = v
	}

	return 0.299*rgb[0] + 0.587*rgb[1] + 0.114*rgb[2], true
}
