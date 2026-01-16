package main

import "github.com/charmbracelet/lipgloss"

// Color palette - modern, clean colors
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#06B6D4") // Cyan
	successColor   = lipgloss.Color("#10B981") // Green
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F9FAFB") // Light gray/white
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	highlightBg    = lipgloss.Color("#374151") // Lighter dark gray
)

// Title/header style - bold, colored
var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(primaryColor).
	Background(bgColor).
	Padding(1, 2).
	MarginBottom(1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor)

// Subtitle style
var subtitleStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Italic(true).
	MarginBottom(1)

// Table header style
var tableHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(textColor).
	Background(primaryColor).
	Padding(0, 1).
	MarginBottom(0)

// Table row style - normal
var tableRowStyle = lipgloss.NewStyle().
	Foreground(textColor).
	Padding(0, 1)

// Table cell style
var tableCellStyle = lipgloss.NewStyle().
	Padding(0, 1)

// Selected row style - highlighted
var selectedRowStyle = lipgloss.NewStyle().
	Foreground(textColor).
	Background(highlightBg).
	Bold(true).
	Padding(0, 1)

// Input field style - for URL input
var inputStyle = lipgloss.NewStyle().
	Foreground(textColor).
	Background(bgColor).
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(secondaryColor)

// Focused input style
var focusedInputStyle = lipgloss.NewStyle().
	Foreground(textColor).
	Background(bgColor).
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor)

// Input label style
var inputLabelStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Bold(true).
	MarginRight(1)

// Help text style - dimmed, for keybindings
var helpStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Italic(true).
	MarginTop(1)

// Help key style
var helpKeyStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Bold(true)

// Help description style
var helpDescStyle = lipgloss.NewStyle().
	Foreground(mutedColor)

// Success message style - green
var successStyle = lipgloss.NewStyle().
	Foreground(successColor).
	Bold(true).
	Padding(0, 1).
	MarginTop(1)

// Error message style - red
var errorStyle = lipgloss.NewStyle().
	Foreground(errorColor).
	Bold(true).
	Padding(0, 1).
	MarginTop(1)

// Warning message style - amber
var warningStyle = lipgloss.NewStyle().
	Foreground(warningColor).
	Bold(true).
	Padding(0, 1)

// Normal text style
var normalTextStyle = lipgloss.NewStyle().
	Foreground(textColor)

// Muted text style
var mutedTextStyle = lipgloss.NewStyle().
	Foreground(mutedColor)

// URL display style - for showing shortened URLs
var urlStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Underline(true)

// Short URL style - highlighted
var shortURLStyle = lipgloss.NewStyle().
	Foreground(primaryColor).
	Bold(true).
	Underline(true)

// Visit count style
var visitCountStyle = lipgloss.NewStyle().
	Foreground(successColor).
	Align(lipgloss.Right)

// Container style - main content wrapper
var containerStyle = lipgloss.NewStyle().
	Padding(1, 2).
	Margin(1)

// Box style - for sections
var boxStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(mutedColor).
	Padding(1, 2).
	MarginBottom(1)

// Focused box style
var focusedBoxStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor).
	Padding(1, 2).
	MarginBottom(1)

// Button style
var buttonStyle = lipgloss.NewStyle().
	Foreground(textColor).
	Background(primaryColor).
	Padding(0, 2).
	MarginRight(1)

// Disabled button style
var disabledButtonStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Background(highlightBg).
	Padding(0, 2).
	MarginRight(1)

// Status bar style
var statusBarStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Background(bgColor).
	Padding(0, 1).
	MarginTop(1)

// Spinner style
var spinnerStyle = lipgloss.NewStyle().
	Foreground(primaryColor)

// Cursor style for text input
var cursorStyle = lipgloss.NewStyle().
	Foreground(primaryColor)

// Placeholder style for text input
var placeholderStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Italic(true)
