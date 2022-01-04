package ui

import "lopper/git"

// errorMsg is a tea.Msg that communicates an error.
type errorMsg struct {
	err error
}

// repositoriesMsg is a tea.Msg that communicates a list of git.Repository to be processed.
type repositoriesMsg struct {
	repositories []git.Repository
}

// inprocessMsg is a tea.Msg that communicates a git.Repository that is currently being processed.
type inprocessMsg struct {
	position   int
	repository git.Repository
}

// completedMsg is a tea.Msg that communicates a git.Repository that has been processed.
type completedMsg struct {
	position int
	branches []string
	errs     []error
}
