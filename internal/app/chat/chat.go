package chat

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
)

const gap = "\n\n"

var uiProc *tea.Program

func Chat(llm *ollamaconnector.Settings) error {
	if log.GetLevel() <= log.DebugLevel {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.SetOutput(f)
		defer f.Close()
	}
	uimodel := initialModel(llm)
	uiProc = tea.NewProgram(&uimodel)
	if _, err := uiProc.Run(); err != nil {

		return err
	}
	return nil
}

type (
	errMsg error
)

type uimodel struct {
	viewport    viewport.Model
	inputs      []string
	answer      string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	ollama      *ollamaconnector.Settings
	uid         string
	mutex       sync.Mutex
	isRunning   bool
	err         error
	db          *database.Knowledge
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
	uid, _ := user.Current()
	db, err := database.New()
	if err != nil {
		log.Warnf("Couldn't create database: %s", err)
	}

	return uimodel{
		textarea:    ta,
		inputs:      []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		ollama:      llm,
		uid:         uid.Username,
		db:          db,
	}
}

func (m *uimodel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *uimodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if len(m.inputs) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.inputs, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.isRunning {
				m.inputs = append(m.inputs, m.senderStyle.Render(m.uid+": ")+m.textarea.Value())
				m.answer = "Kowalski: "
				m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(
					strings.Join(m.inputs, "\n")))
				m.TalkLLMBackground(m.textarea.Value())
				m.textarea.Reset()
				m.viewport.GotoBottom()
			}
		}
	case LLMAns:
		m.answer += string(msg)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(
			strings.Join(append(m.inputs, m.answer), "\n")))

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *uimodel) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

type LLMAns string

func (m *uimodel) TalkLLMBackground(msg string) error {
	if m.isRunning {
		return nil
	}
	m.mutex.Lock()
	m.isRunning = true
	m.mutex.Unlock()
	prompt, err := m.db.GetContext(msg, []string{}, m.ollama.ContextLength)
	if err != nil {
		m.err = err
		fmt.Println("An errror occured", err)
		return nil
	}
	ch := make(chan *ollamaconnector.TaskResponse)
	go m.ollama.SendTaskStream(prompt, ch)
	go func() {
		for resp := range ch {
			uiProc.Send(LLMAns(resp.Response))
		}
		m.mutex.Lock()
		m.inputs = append(m.inputs, m.answer)
		m.isRunning = false
		m.answer = ""
		m.mutex.Unlock()
	}()
	return nil
}
