package ui

import (
	"golang.org/x/sync/semaphore"
)

// Option is a function that is used to update the Model.
type Option func(m *Model)

// Path sets the path of the file or directory to be operated on.
func Path(path string) Option {
	return func(m *Model) {
		m.path = path
	}
}

// ProtectedBranches sets the protected branches of the repository.
func ProtectedBranches(protectedBranches []string) Option {
	return func(m *Model) {
		m.protectedBranches = protectedBranches
	}
}

// Concurrency sets the number of repositories to be processed in parallel.
func Concurrency(concurrency int) Option {
	return func(m *Model) {
		m.startProcessMsgs = make(chan inprocessMsg, concurrency)
		m.completedMsgs = make(chan completedMsg, concurrency)
		m.semaphore = semaphore.NewWeighted(int64(concurrency))
	}
}

// DryRun sets does not delete any branches.
func DryRun(dryRun bool) Option {
	return func(m *Model) {
		m.dryRun = dryRun
	}
}
