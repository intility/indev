package ux

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

var (
	StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

	StyleSuccessLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("10"))
	StyleWarningLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11"))
	StyleErrorLabel   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("9"))
	StyleInfoLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("14"))
)

func Fsuccessf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, StyleSuccess.Render("success: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

//goland:noinspection GoUnusedExportedFunction
func Finfof(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, StyleInfo.Render("info: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

//goland:noinspection GoUnusedExportedFunction
func Fwarningf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, StyleWarning.Render("warning: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Ferrorf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, StyleError.Render("error: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Fprintf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
