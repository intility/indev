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
	InputTypeToggle
	InputTypeSelect
)

type Input struct {
	ID          string
	Placeholder string
	Type        InputType
	Limit       int
	Validator   func(string) error
	Options     []string                             // For select and toggle types
	DependsOn   string                               // ID of field this depends on
	ShowWhen    func(answers map[string]Answer) bool // Conditional visibility
}

type Answer struct {
	Value string
}

// inputField represents any type of input field (text, select, toggle)
type inputField interface {
	Focus() tea.Cmd
	Blur()
	Value() string
	SetValue(string)
	Update(tea.Msg) (inputField, tea.Cmd)
	View() string
}

// textField wraps textinput.Model to implement inputField
type textField struct {
	model textinput.Model
}

func (t *textField) Focus() tea.Cmd    { return t.model.Focus() }
func (t *textField) Blur()             { t.model.Blur() }
func (t *textField) Value() string     { return t.model.Value() }
func (t *textField) SetValue(v string) { t.model.SetValue(v) }
func (t *textField) Update(msg tea.Msg) (inputField, tea.Cmd) {
	newModel, cmd := t.model.Update(msg)
	t.model = newModel
	return t, cmd
}
func (t *textField) View() string { return t.model.View() }

// selectField implements a simple select/toggle field
type selectField struct {
	id          string
	placeholder string
	options     []string
	selected    int
	focused     bool
	isToggle    bool
	style       lipgloss.Style
}

func (s *selectField) Focus() tea.Cmd {
	s.focused = true
	return nil
}

func (s *selectField) Blur() {
	s.focused = false
}

func (s *selectField) Value() string {
	if len(s.options) == 0 {
		return ""
	}
	return s.options[s.selected]
}

func (s *selectField) SetValue(v string) {
	for i, opt := range s.options {
		if opt == v {
			s.selected = i
			return
		}
	}
}

func (s *selectField) Update(msg tea.Msg) (inputField, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "enter":
			if s.isToggle {
				// Toggle between yes/no
				s.selected = (s.selected + 1) % len(s.options)
			}
		case "left", "h":
			if s.selected > 0 {
				s.selected--
			}
		case "right", "l":
			if s.selected < len(s.options)-1 {
				s.selected++
			}
		}
	}
	return s, nil
}

func (s *selectField) View() string {
	var b strings.Builder

	if s.isToggle {
		// Render as toggle (e.g., "[x] Yes  [ ] No")
		b.WriteString(s.placeholder + ": ")
		for i, opt := range s.options {
			if i == s.selected {
				b.WriteString(s.style.Render("[●] " + opt))
			} else {
				b.WriteString(s.style.Render("[ ] " + opt))
			}
			if i < len(s.options)-1 {
				b.WriteString("  ")
			}
		}
	} else {
		// Render as select list
		b.WriteString(s.placeholder + ": ")
		b.WriteString(s.style.Render("< " + s.options[s.selected] + " >"))
	}

	return b.String()
}

type model struct {
	focusIndex    int
	items         []Input
	fields        []inputField // Changed from []textinput.Model to []inputField
	visibleFields []int        // Indices of currently visible fields
	answers       map[string]Answer
	cursorMode    cursor.Mode
	state         *State

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

	var cmds []tea.Cmd
	for _, field := range m.fields {
		if tf, ok := field.(*textField); ok {
			cmds = append(cmds, tf.model.Cursor.SetMode(m.cursorMode))
		}
	}

	return tea.Batch(cmds...)
}

func (m *model) updateVisibility() {
	// Update which fields are visible based on current answers
	m.visibleFields = []int{}

	for i, item := range m.items {
		// Check if this field should be visible
		if item.ShowWhen == nil || item.ShowWhen(m.answers) {
			m.visibleFields = append(m.visibleFields, i)
		}
	}
}

func (m *model) handleNavigation(s string) tea.Cmd {
	// Update answers for current field before navigation
	if m.focusIndex < len(m.visibleFields) {
		fieldIdx := m.visibleFields[m.focusIndex]
		m.answers[m.items[fieldIdx].ID] = Answer{
			Value: m.fields[fieldIdx].Value(),
		}
		// Update visibility after answer changes
		m.updateVisibility()
	}

	// Did the user press enter while the submit button was focused?
	// If so, exit.
	if s == "enter" && m.focusIndex == len(m.visibleFields) {
		// Collect all answers from visible fields
		for _, idx := range m.visibleFields {
			m.answers[m.items[idx].ID] = Answer{
				Value: m.fields[idx].Value(),
			}
		}

		m.state.Complete()
		return tea.Quit
	}

	// Cycle indexes through visible fields
	if s == "up" || s == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}

	if m.focusIndex > len(m.visibleFields) {
		m.focusIndex = 0
	} else if m.focusIndex < 0 {
		m.focusIndex = len(m.visibleFields)
	}

	var cmds []tea.Cmd

	// Update focus for visible fields
	for i, visIdx := range m.visibleFields {
		field := m.fields[visIdx]

		if i == m.focusIndex {
			// Set focused state
			cmds = append(cmds, field.Focus())

			// Update style for text fields
			if tf, ok := field.(*textField); ok {
				tf.model.PromptStyle = m.focusedStyle
				tf.model.TextStyle = m.focusedStyle
			} else if sf, ok := field.(*selectField); ok {
				sf.style = m.focusedStyle
			}
		} else {
			// Remove focused state
			field.Blur()

			// Update style for text fields
			if tf, ok := field.(*textField); ok {
				tf.model.PromptStyle = m.noStyle
				tf.model.TextStyle = m.noStyle
			} else if sf, ok := field.(*selectField); ok {
				sf.style = m.blurredStyle
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update only visible fields
	for _, idx := range m.visibleFields {
		field, cmd := m.fields[idx].Update(msg)
		m.fields[idx] = field
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	// Only show visible fields
	for i, idx := range m.visibleFields {
		b.WriteString(m.fields[idx].View())

		if i < len(m.visibleFields)-1 {
			b.WriteRune('\n')
		}
	}

	button := &m.blurredButton
	if m.focusIndex == len(m.visibleFields) {
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
		items:         inputs,
		cursorMode:    cursor.CursorStatic,
		fields:        make([]inputField, len(inputs)),
		visibleFields: []int{},
		answers:       make(map[string]Answer),
		state:         &State{FinishReason: FinishReasonCompleted},
		focusIndex:    0,

		focusedStyle:  focusedStyle,
		blurredStyle:  blurredStyle,
		cursorStyle:   focusedStyle.Copy(),
		noStyle:       lipgloss.NewStyle(),
		helpStyle:     blurredStyle.Copy().Foreground(lipgloss.Color("244")),
		focusedButton: focusedStyle.Copy().Render("[ Submit ]"),
		blurredButton: fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit")),
	}

	// Create appropriate field type for each input
	for i, input := range inputs {
		switch input.Type {
		case InputTypeToggle:
			// For toggle, default to yes/no options if not specified
			options := input.Options
			if len(options) == 0 {
				options = []string{"no", "yes"}
			}
			m.fields[i] = &selectField{
				id:          input.ID,
				placeholder: input.Placeholder,
				options:     options,
				selected:    0,
				isToggle:    true,
				style:       blurredStyle,
			}

		case InputTypeSelect:
			m.fields[i] = &selectField{
				id:          input.ID,
				placeholder: input.Placeholder,
				options:     input.Options,
				selected:    0,
				isToggle:    false,
				style:       blurredStyle,
			}

		default: // InputTypeText and InputTypePassword
			t := textinput.New()
			t.Cursor.Style = m.cursorStyle
			t.CharLimit = 32
			t.Width = 50
			t.Placeholder = input.Placeholder

			if input.Validator != nil {
				t.Validate = input.Validator
			}

			if input.Type == InputTypePassword {
				t.EchoMode = textinput.EchoPassword
				t.EchoCharacter = '•'
			}

			if input.Limit > 0 {
				t.CharLimit = input.Limit
			}

			m.fields[i] = &textField{model: t}
		}
	}

	// Initialize visibility
	m.updateVisibility()

	// Focus the first visible field
	if len(m.visibleFields) > 0 {
		firstField := m.fields[m.visibleFields[0]]
		firstField.Focus()

		if tf, ok := firstField.(*textField); ok {
			tf.model.PromptStyle = focusedStyle
			tf.model.TextStyle = focusedStyle
		} else if sf, ok := firstField.(*selectField); ok {
			sf.style = focusedStyle
		}
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
