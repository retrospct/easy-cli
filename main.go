package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"easy-cli/client"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


type tickMsg struct{}
type errMsg error
type model struct {
	// cursor   int
	// choices  []string
	// selected map[int]struct{}
	url 			string
	spinner  	spinner.Model
	textInput textinput.Model
	err       error
}

func initialModel() model {
	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Text input
	ti := textinput.New()
	ti.Placeholder = "<enter URL>"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return model{
		// choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		// // A map which indicates which choices are selected. We're using
		// // the  map like a mathematical set. The keys refer to the indexes
		// // of the `choices` slice, above.
		// selected: make(map[int]struct{}),
		spinner:  s,
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.url, m.err = shorten(m.textInput.Value())
		// case "up", "k":
		// 	if m.cursor > 0 {
		// 		m.cursor--
		// 	}
		// case "down", "j":
		// 	if m.cursor < len(m.choices)-1 {
		// 		m.cursor++
		// 	}
		// case "enter", " ":
		// 	_, ok := m.selected[m.cursor]
		// 	if ok {
		// 		delete(m.selected, m.cursor)
		// 	} else {
		// 		m.selected[m.cursor] = struct{}{}
		// 	}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.spinner, cmd = m.spinner.Update(msg)
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Text input
	s := fmt.Sprintf(
		"URL to shorten: %s\n\n",
		m.textInput.View(),
	)
	// s += fmt.Sprintf("\n%s\n", m.url)
	
	// Spinner
	s += fmt.Sprintf("\n\n   %s Loading forever...\n\n", m.spinner.View())
	
	// for i, choice := range m.choices {
	// 	cursor := " "
	// 	if m.cursor == i {
	// 		cursor = ">"
	// 	}
		
	// 	checked := " "
	// 	if _, ok := m.selected[i]; ok {
	// 		checked = "x"
	// 	}
		
	// 	s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	// }

	s += "\nPress esc to quit.\n"

	return s
}

func shorten(url string) (string, error) {
	// Create a new client with the default BaseURL
	client, err := client.New(
		client.Environment("production"),
		client.WithAuth(os.Getenv("SHORTEN_API_KEY")),
	)
	if err != nil {
			panic(err)
	}

	// Timeout if the request takes more than 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the Shorten functon in the URL service
	resp, err := client.Url.Shorten(
			ctx,
			client.UrlShortenParams{ URL: os.Args[1] },
	)
	if err != nil {
			// Check the error returned
			if err, ok := err.(*client.APIError); ok {
					switch err.Code {
					case client.ErrUnauthenticated:
							fmt.Println("SHORTEN_API_KEY was invalid, please check your environment")
							os.Exit(1)
					case client.ErrAlreadyExists:
							fmt.Println("The URL you provided was already shortened")
							os.Exit(0)
					}
			}
			panic(err) // if here then something has gone wrong in an unexpected way
	}

	shortUrl := fmt.Sprintf("https://short.encr.app/%s", resp.ID)
	// Return the shortened URL
	return shortUrl, nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}