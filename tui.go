package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tuiModel struct {
	tasks    []Task
	cursor   int
	selected int // original index chosen by enter, -1 if none
	client   *memoClient
}

func newTUIModel(tasks []Task, client *memoClient) tuiModel {
	return tuiModel{
		tasks:    tasks,
		cursor:   0,
		selected: -1,
		client:   client,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	s := ""
	for i, task := range m.tasks {
		cursor := "  "
		if i == m.cursor {
			cursor = "→ "
		}

		desc := task.Description
		if i == 0 {
			desc = fmt.Sprintf("%s (working for %s)", desc, formatDuration(time.Since(task.StartedAt)))
		}

		s += fmt.Sprintf("%s%s\n", cursor, desc)
	}
	return s
}

func runTUI(client *memoClient) {
	stack, err := client.FetchStack()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	if stack.Len() < 2 {
		client.Stack()
		return
	}

	m := newTUIModel(stack.Tasks, client)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	final := finalModel.(tuiModel)
	if final.selected > 0 {
		// Build reorder: move selected to top, shift others down
		order := make([]int, len(final.tasks))
		order[0] = final.selected
		pos := 1
		for i := range final.tasks {
			if i != final.selected {
				order[pos] = i
				pos++
			}
		}
		if err := client.Reorder(order); err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Printf("Paused: %s\n", final.tasks[0].Description)
		fmt.Printf("Resuming: %s\n", final.tasks[final.selected].Description)
	}
}
