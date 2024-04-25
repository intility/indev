package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FinishReason int

const (
	FinishReasonCompleted FinishReason = iota
	FinishReasonCancelled
)

type State struct {
	FinishReason FinishReason
}

func (s *State) Complete() {
	s.FinishReason = FinishReasonCompleted
}

func (s *State) Cancel() {
	s.FinishReason = FinishReasonCancelled
}

type InputType int

const (
	InputTypeText InputType = iota
	InputTypePassword
)

type Input struct {
	ID          string
	Placeholder string
	Type        InputType
	Limit       int
	Validator   func(string) error
}

type Answer struct {
	Value string
}

type model struct {
	focusIndex int
	items      []Input
	inputs     []textinput.Model
	answers    map[string]Answer
	cursorMode cursor.Mode
	state      *State

	focusedStyle  lipgloss.Style
	blurredStyle  lipgloss.Style
	cursorStyle   lipgloss.Style
	noStyle       lipgloss.Style
	helpStyle     lipgloss.Style
	focusedButton string
	blurredButton string
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) { //nolint: gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.state.Cancel()
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			cmd := m.handleCursorMode()
			return m, cmd

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			cmd := m.handleNavigation(msg.String())
			return m, cmd
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) handleCursorMode() tea.Cmd {
	m.cursorMode++

	if m.cursorMode > cursor.CursorHide {
		m.cursorMode = cursor.CursorBlink
	}

	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
	}

	return tea.Batch(cmds...)
}

func (m *model) handleNavigation(s string) tea.Cmd {
	// Did the user press enter while the submit button was focused?
	// If so, exit.
	if s == "enter" && m.focusIndex == len(m.inputs) {
		// construct answers based on the current state of the inputs
		// matching them with the index of the items
		for i := range m.items {
			m.answers[m.items[i].ID] = Answer{
				Value: m.inputs[i].Value(),
			}
		}

		m.state.Complete()

		return tea.Quit
	}

	// Cycle indexes
	if s == "up" || s == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.inputs) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs)
	}

	cmds := make([]tea.Cmd, len(m.inputs))

	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = m.focusedStyle
			m.inputs[i].TextStyle = m.focusedStyle

			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = m.noStyle
		m.inputs[i].TextStyle = m.noStyle
	}

	return tea.Batch(cmds...)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())

		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &m.blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &m.focusedButton
	}

	_, _ = fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(m.helpStyle.Render("(esc to cancel)"))

	return b.String()
}

type Wizard struct {
	model model
}

func NewWizard(inputs []Input) *Wizard {
	var (
		focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	)

	m := model{
		items:      inputs,
		cursorMode: cursor.CursorStatic,
		inputs:     make([]textinput.Model, len(inputs)),
		answers:    make(map[string]Answer),
		state:      &State{FinishReason: FinishReasonCompleted},
		focusIndex: 0,

		focusedStyle:  focusedStyle,
		blurredStyle:  blurredStyle,
		cursorStyle:   focusedStyle.Copy(),
		noStyle:       lipgloss.NewStyle(),
		helpStyle:     blurredStyle.Copy().Foreground(lipgloss.Color("244")),
		focusedButton: focusedStyle.Copy().Render("[ Submit ]"),
		blurredButton: fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit")),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = m.cursorStyle
		t.CharLimit = 32

		t.Placeholder = inputs[i].Placeholder

		if inputs[i].Validator != nil {
			t.Validate = inputs[i].Validator
		}

		if inputs[i].Type == InputTypePassword {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		if inputs[i].Limit > 0 {
			t.CharLimit = inputs[i].Limit
		}

		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}

		m.inputs[i] = t
	}

	return &Wizard{
		model: m,
	}
}

type Result struct {
	finishReason FinishReason
	answers      map[string]Answer
}

func (w *Result) MustGetValue(id string) string {
	answer, ok := w.answers[id]
	if !ok {
		panic(fmt.Sprintf("answer with id %s not found", id))
	}

	return answer.Value
}

func (w *Result) Cancelled() bool {
	return w.finishReason == FinishReasonCancelled
}

func (w *Wizard) Run() (Result, error) {
	var result Result

	p := tea.NewProgram(w.model)
	if _, err := p.Run(); err != nil {
		return result, fmt.Errorf("error running wizard: %w", err)
	}

	result = Result{
		finishReason: w.model.state.FinishReason,
		answers:      w.model.answers,
	}

	return result, nil
}
