package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
)

const gap = "\n\n"

func Chat(llm *ollamaconnector.Settings) error {
	uimodel := initialModel(llm)
	p := tea.NewProgram(uimodel)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

type (
	errMsg error
)

type uimodel struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	ollama      *ollamaconnector.Settings
	err         error
}

func initialModel(llm *ollamaconnector.Settings) uimodel {
	ta := textarea.New()
	ta.Placeholder = "Type CTR-C or ESC to quit..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to a system configuration prompt.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return uimodel{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		ollama:      llm,
	}
}

func (m uimodel) Init() tea.Cmd {
	return textarea.Blink
}

func (m uimodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			resp, err := m.ollama.TalkToOllama(
				[]ollamaconnector.ChatMessage{{
					Role:    "user",
					Content: m.textarea.Value(),
				}})
			if err != nil {
				m.err = err
				fmt.Println("An errror occured", err)
				return m, nil
			}
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value(),
				"Kowalski: "+resp.Message.Content)
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(
				strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m uimodel) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
