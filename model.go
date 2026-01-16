package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// viewState represents the current view of the TUI
type viewState int

const (
	listView viewState = iota
	createView
	deleteConfirmView
)

// model is the main Bubble Tea model for the TUI
type model struct {
	provider  URLProvider     // Data provider (local App or remote APIClient)
	urls      []CustomURL     // Current URL list
	table     table.Model     // bubbles table component
	textInput textinput.Model // For URL input
	state     viewState       // Current view
	cursor    int             // Selected row
	message   string          // Status/success message
	err       error           // Error state
	width     int             // Terminal width
	height    int             // Terminal height
	isRemote  bool            // True if using remote API
}

// Custom messages for async operations
type urlsLoadedMsg struct {
	urls []CustomURL
}

type urlCreatedMsg struct {
	url *CustomURL
}

type urlDeletedMsg struct {
	id uint
}

type errMsg struct {
	err error
}

// initialModel creates the initial model with default values
func initialModel(provider URLProvider, isRemote bool) model {
	ti := textinput.New()
	ti.Placeholder = "https://example.com/long-url-to-shorten"
	ti.Focus()
	ti.CharLimit = 2048
	ti.Width = 60
	ti.PromptStyle = inputLabelStyle
	ti.TextStyle = normalTextStyle
	ti.PlaceholderStyle = placeholderStyle
	ti.Cursor.Style = cursorStyle

	m := model{
		provider:  provider,
		urls:      []CustomURL{},
		textInput: ti,
		state:     listView,
		cursor:    0,
		message:   "",
		err:       nil,
		width:     80,
		height:    24,
		isRemote:  isRemote,
	}

	m.table = buildTable(m.urls, m.width)
	return m
}

// Init initializes the model and returns a command to load URLs
func (m model) Init() tea.Cmd {
	return m.loadURLs
}

// loadURLs fetches all URLs from the provider
func (m model) loadURLs() tea.Msg {
	urls, err := m.provider.GetAllURLs()
	if err != nil {
		return errMsg{err: err}
	}
	return urlsLoadedMsg{urls: urls}
}

// createURL creates a new short URL via the provider
func (m model) createURL(originalURL string) tea.Cmd {
	return func() tea.Msg {
		url, err := m.provider.CreateURL(originalURL)
		if err != nil {
			return errMsg{err: err}
		}
		return urlCreatedMsg{url: url}
	}
}

// deleteURL deletes a URL via the provider
func (m model) deleteURL(id uint) tea.Cmd {
	return func() tea.Msg {
		err := m.provider.DeleteURL(id)
		if err != nil {
			return errMsg{err: err}
		}
		return urlDeletedMsg{id: id}
	}
}

// Update handles all messages and user input
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = buildTable(m.urls, m.width)
		return m, nil

	case urlsLoadedMsg:
		m.urls = msg.urls
		m.table = buildTable(m.urls, m.width)
		m.message = fmt.Sprintf("Loaded %d URLs", len(m.urls))
		m.err = nil
		return m, nil

	case urlCreatedMsg:
		m.message = fmt.Sprintf("Created short URL: %s", msg.url.ShortURL)
		m.state = listView
		m.textInput.Reset()
		m.err = nil
		return m, m.loadURLs

	case urlDeletedMsg:
		m.message = "URL deleted successfully"
		m.state = listView
		m.err = nil
		return m, m.loadURLs

	case errMsg:
		m.err = msg.err
		m.message = ""
		return m, nil

	case clipboardMsg:
		if msg.success {
			m.message = fmt.Sprintf("Copied to clipboard: %s", msg.text)
		} else {
			m.message = fmt.Sprintf("Could not copy to clipboard (xclip not installed?): %s", msg.text)
		}
		return m, nil

	case tea.KeyMsg:
		// Clear error on any key press
		if m.err != nil && msg.String() != "ctrl+c" && msg.String() != "q" {
			m.err = nil
		}

		switch m.state {
		case listView:
			return m.updateListView(msg)
		case createView:
			return m.updateCreateView(msg)
		case deleteConfirmView:
			return m.updateDeleteConfirmView(msg)
		}
	}

	return m, cmd
}

// updateListView handles input for the list view
func (m model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "n":
		m.state = createView
		m.textInput.Focus()
		m.message = ""
		return m, nil

	case "d":
		if len(m.urls) > 0 {
			m.state = deleteConfirmView
			m.message = ""
		}
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.table.SetCursor(m.cursor)
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.urls)-1 {
			m.cursor++
			m.table.SetCursor(m.cursor)
		}
		return m, nil

	case "enter":
		if len(m.urls) > 0 && m.cursor < len(m.urls) {
			shortURL := m.urls[m.cursor].ShortURL
			return m, copyToClipboardCmd(shortURL)
		}
		return m, nil
	}

	// Update table component
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	m.cursor = m.table.Cursor()
	return m, cmd
}

// updateCreateView handles input for the create URL view
func (m model) updateCreateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = listView
		m.textInput.Reset()
		m.message = ""
		return m, nil

	case "enter":
		url := strings.TrimSpace(m.textInput.Value())
		if url != "" {
			return m, m.createURL(url)
		}
		m.err = fmt.Errorf("URL cannot be empty")
		return m, nil
	}

	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updateDeleteConfirmView handles input for the delete confirmation view
func (m model) updateDeleteConfirmView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.urls) {
			id := m.urls[m.cursor].ID
			return m, m.deleteURL(id)
		}
		m.state = listView
		return m, nil

	case "n", "N", "esc":
		m.state = listView
		m.message = "Delete cancelled"
		return m, nil
	}

	return m, nil
}

// View renders the current view
func (m model) View() string {
	var b strings.Builder

	switch m.state {
	case listView:
		b.WriteString(m.renderListView())
	case createView:
		b.WriteString(m.renderCreateView())
	case deleteConfirmView:
		b.WriteString(m.renderDeleteConfirmView())
	}

	return b.String()
}

// renderListView renders the main list view with the URL table
func (m model) renderListView() string {
	var b strings.Builder

	// Title
	titleText := "ShortMe - URL Shortener"
	if m.isRemote {
		titleText += " [Remote]"
	}
	title := titleStyle.Render(titleText)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Table
	if len(m.urls) == 0 {
		b.WriteString(mutedTextStyle.Render("No URLs found. Press 'n' to create one."))
		b.WriteString("\n")
	} else {
		b.WriteString(m.table.View())
		b.WriteString("\n")
	}

	// Status message or error
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.message != "" {
		b.WriteString(successStyle.Render(m.message))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(formatHelp(listView))

	return containerStyle.Render(b.String())
}

// renderCreateView renders the create URL form
func (m model) renderCreateView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Create Short URL")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Input label
	b.WriteString(inputLabelStyle.Render("Enter URL to shorten:"))
	b.WriteString("\n")

	// Text input
	b.WriteString(focusedInputStyle.Render(m.textInput.View()))
	b.WriteString("\n")

	// Error message
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(formatHelp(createView))

	return containerStyle.Render(b.String())
}

// renderDeleteConfirmView renders the delete confirmation dialog
func (m model) renderDeleteConfirmView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Confirm Delete")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Confirmation message
	if m.cursor < len(m.urls) {
		url := m.urls[m.cursor]
		b.WriteString(warningStyle.Render("Are you sure you want to delete this URL?"))
		b.WriteString("\n\n")
		b.WriteString(normalTextStyle.Render("Short URL: "))
		b.WriteString(shortURLStyle.Render(url.ShortURL))
		b.WriteString("\n")
		b.WriteString(normalTextStyle.Render("Original:  "))
		b.WriteString(urlStyle.Render(truncateString(url.OldURL, 60)))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(formatHelp(deleteConfirmView))

	return containerStyle.Render(b.String())
}

// buildTable creates a table model from the URL list
func buildTable(urls []CustomURL, width int) table.Model {
	columns := []table.Column{
		{Title: "Short Code", Width: 20},
		{Title: "Original URL", Width: 40},
		{Title: "Visits", Width: 8},
		{Title: "Created", Width: 12},
	}

	// Adjust column widths based on terminal width
	if width > 120 {
		columns[1].Width = 60
	} else if width < 80 {
		columns[0].Width = 15
		columns[1].Width = 30
	}

	rows := make([]table.Row, len(urls))
	for i, url := range urls {
		// Extract just the short code from the full URL
		shortCode := url.ShortURL
		if idx := strings.LastIndex(shortCode, "/"); idx != -1 {
			shortCode = shortCode[idx+1:]
		}

		rows[i] = table.Row{
			shortCode,
			truncateString(url.OldURL, columns[1].Width),
			fmt.Sprintf("%d", url.Visits),
			url.CreatedAt.Format("2006-01-02"),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = selectedRowStyle
	s.Cell = tableCellStyle
	t.SetStyles(s)

	return t
}

// formatHelp returns the help text for the current view state
func formatHelp(state viewState) string {
	var keys []string

	switch state {
	case listView:
		keys = []string{
			helpKeyStyle.Render("n") + helpDescStyle.Render(" new"),
			helpKeyStyle.Render("d") + helpDescStyle.Render(" delete"),
			helpKeyStyle.Render("j/k") + helpDescStyle.Render(" navigate"),
			helpKeyStyle.Render("enter") + helpDescStyle.Render(" copy"),
			helpKeyStyle.Render("q") + helpDescStyle.Render(" quit"),
		}
	case createView:
		keys = []string{
			helpKeyStyle.Render("enter") + helpDescStyle.Render(" submit"),
			helpKeyStyle.Render("esc") + helpDescStyle.Render(" cancel"),
		}
	case deleteConfirmView:
		keys = []string{
			helpKeyStyle.Render("y") + helpDescStyle.Render(" yes, delete"),
			helpKeyStyle.Render("n/esc") + helpDescStyle.Render(" no, cancel"),
		}
	}

	return helpStyle.Render(strings.Join(keys, "  |  "))
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
