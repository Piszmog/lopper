package ui

import "github.com/charmbracelet/lipgloss"

// Symbols

var symbolX = "✘"
var symbolCheck = "✔"
var symbolBranch = "├"
var symbolLeaf = "└"

// Color Styles

var spinnerColor = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var completedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
var grayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
