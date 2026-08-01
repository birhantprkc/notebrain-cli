package cmd

import (
	"image/color"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

var (
	stylesOnce     sync.Once
	colorAccent    color.Color
	colorMuted     color.Color
	colorGood      color.Color
	colorWarn      color.Color
	colorError     color.Color
	headerStyle    lipgloss.Style
	scoreStyle     lipgloss.Style
	warnScoreStyle lipgloss.Style
	extraStyle     lipgloss.Style
	hintStyle      lipgloss.Style
	rankStyle      lipgloss.Style
	boxStyle       lipgloss.Style
	titleStyle     lipgloss.Style
	metaStyle      lipgloss.Style
	labelStyle     lipgloss.Style
	warnBoldStyle  lipgloss.Style
	errorStyle     lipgloss.Style
)

// stdoutAllowsColor reports whether ANSI colors may be emitted on stdout.
// Colors are disabled when stdout is not a terminal, when TERM=dumb, or when
// NO_COLOR is set, so piped and redirected output stays machine-clean.
func stdoutAllowsColor() bool {
	return term.IsTerminal(os.Stdout.Fd()) && os.Getenv("TERM") != "dumb" && os.Getenv("NO_COLOR") == ""
}

func initStyles() {
	stylesOnce.Do(func() {
		hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
		lightDark := lipgloss.LightDark(hasDarkBG)

		colorAccent = lightDark(lipgloss.Color("#534AB7"), lipgloss.Color("#AFA9EC")) // purple
		colorMuted = lightDark(lipgloss.Color("#888780"), lipgloss.Color("#B4B2A9"))  // gray
		colorGood = lightDark(lipgloss.Color("#0F6E56"), lipgloss.Color("#5DCAA5"))   // teal
		colorWarn = lightDark(lipgloss.Color("#C4841D"), lipgloss.Color("#F5A623"))   // amber/orange
		colorError = lightDark(lipgloss.Color("#B3261E"), lipgloss.Color("#FF6B6B"))  // red

		if !stdoutAllowsColor() {
			// Plain rendering for piped/redirected stdout: keep layout
			// (borders, padding, width) but drop colors and text effects so
			// that NO_COLOR, TERM=dumb, and non-TTY output stay clean.
			headerStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				PaddingBottom(1)
			rankStyle = lipgloss.NewStyle().
				Width(3).
				Align(lipgloss.Right)
			boxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1)
			scoreStyle = lipgloss.NewStyle()
			warnScoreStyle = lipgloss.NewStyle()
			extraStyle = lipgloss.NewStyle()
			hintStyle = lipgloss.NewStyle()
			titleStyle = lipgloss.NewStyle()
			metaStyle = lipgloss.NewStyle()
			labelStyle = lipgloss.NewStyle()
			warnBoldStyle = lipgloss.NewStyle()
			errorStyle = lipgloss.NewStyle()
			return
		}

		headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorAccent).
			Foreground(colorAccent).
			Bold(true).
			PaddingBottom(1)

		scoreStyle = lipgloss.NewStyle().
			Foreground(colorGood)

		warnScoreStyle = lipgloss.NewStyle().
			Foreground(colorWarn)

		extraStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

		hintStyle = lipgloss.NewStyle().
			Foreground(colorWarn).
			Italic(true)

		rankStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(3).
			Align(lipgloss.Right)

		boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

		metaStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

		labelStyle = lipgloss.NewStyle().
			Bold(true)

		warnBoldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarn)

		errorStyle = lipgloss.NewStyle().
			Foreground(colorError)
	})
}

func scoreStyleFor(score float64) lipgloss.Style {
	initStyles()
	switch {
	case score >= 0.75:
		return scoreStyle // teal — strong match
	case score >= 0.50:
		return warnScoreStyle // amber — moderate match
	default:
		return extraStyle // gray — weak match
	}
}
