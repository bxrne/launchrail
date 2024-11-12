package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/bxrne/launchrail/pkg/ork"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	statePickingOrk state = iota
	statePickingThrust
	stateAssembling
	stateComplete
	stateError
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF06B7")).
			Bold(true).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			MarginLeft(2)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			MarginLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			MarginLeft(2)

	filePickerStyle = lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1)
)

type model struct {
	filePicker filepicker.Model
	spinner    spinner.Model
	state      state
	err        error
	assembly   *entities.Assembly
	orkFile    string
	thrustFile string
}

func initialModel() model {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.Getwd()
	fp.AllowedTypes = []string{".ork"}
	fp.ShowHidden = false
	fp.Height = 10

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		filePicker: fp,
		spinner:    s,
		state:      statePickingOrk,
	}
}

type assemblyMsg struct {
	assembly *entities.Assembly
	err      error
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state == statePickingThrust {
				m.state = statePickingOrk
				m.filePicker.AllowedTypes = []string{".ork"}
				m.filePicker.CurrentDirectory = filepath.Dir(m.orkFile)
				return m, m.filePicker.Init()
			}
		}

	case assemblyMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
			return m, nil
		}
		m.assembly = msg.assembly
		m.state = stateComplete
		return m, nil

	case tickMsg:
		if m.state == stateAssembling {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	switch m.state {
	case statePickingOrk, statePickingThrust:
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			switch m.state {
			case statePickingOrk:
				if filepath.Ext(path) == ".ork" {
					m.orkFile = path
					m.state = statePickingThrust
					m.filePicker.AllowedTypes = []string{".eng"}
					m.filePicker.CurrentDirectory = filepath.Dir(path)
					cmds = append(cmds, m.filePicker.Init())
				}
			case statePickingThrust:
				if filepath.Ext(path) == ".eng" {
					m.thrustFile = path
					m.state = stateAssembling
					return m, tea.Batch(
						m.spinner.Tick,
						m.performAssembly,
					)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 Launchrail (dev)"))
	b.WriteString("\n\n")

	switch m.state {
	case statePickingOrk:
		b.WriteString(subtitleStyle.Render("Select OpenRocket file (.ork):"))
		b.WriteString("\n")
		b.WriteString(filePickerStyle.Render(m.filePicker.View()))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Press q to quit, enter to select"))

	case statePickingThrust:
		b.WriteString(subtitleStyle.Render("OpenRocket file: " + m.orkFile))
		b.WriteString("\n\n")
		b.WriteString(subtitleStyle.Render("Select thrust curve for Solid Motor (.eng):"))
		b.WriteString("\n")
		b.WriteString(filePickerStyle.Render(m.filePicker.View()))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Press esc to go back, q to quit, enter to select"))

	case stateAssembling:
		b.WriteString(fmt.Sprintf("%s Assembling rocket...", m.spinner.View()))

	case stateComplete:
		b.WriteString(subtitleStyle.Render("Assembly Complete! 🎉"))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render("Details:"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(m.assembly.Info()))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render("Press q to quit"))

	case stateError:
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Press q to quit"))
	}

	return b.String()
}

func (m model) performAssembly() tea.Msg {
	log := logger.GetLogger()

	orkData, err := ork.Decompress(m.orkFile)
	if err != nil {
		return assemblyMsg{
			err: fmt.Errorf("failed to decompress OpenRocket file: %w", err),
		}
	}

	assembly, err := entities.NewRocket(*orkData, m.thrustFile)
	if err != nil {
		return assemblyMsg{
			err: fmt.Errorf("failed to create rocket assembly: %w", err),
		}
	}

	log.Infof("Assembly complete for %s", assembly.Info())
	return assemblyMsg{
		assembly: assembly,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
