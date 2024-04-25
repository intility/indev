package ux

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

func Fsuccess(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, styleSuccess.Render("success: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Finfo(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, styleInfo.Render("info: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Fwarning(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, styleWarning.Render("warning: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Ferror(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprint(w, styleError.Render("error: "))
	_, _ = fmt.Fprintf(w, format, a...)
}

func Fprint(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
