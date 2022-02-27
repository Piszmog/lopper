package ui

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/semaphore"
	"lopper/git"
	"lopper/utils"
	"os"
	"path/filepath"
	"strings"
)

// Model is the model for the UI.
type Model struct {
	// configuration properties
	path              string
	protectedBranches []string
	dryRun            bool

	// state properties
	repositories    []git.Repository
	states          map[int]state
	deletedBranches map[int][]string
	errMessages     map[int][]error

	// view properties
	spinner  spinner.Model
	viewport viewport.Model
	builder  strings.Builder

	// other properties
	ready            bool
	semaphore        *semaphore.Weighted
	startProcessMsgs chan inprocessMsg
	completedMsgs    chan completedMsg
	err              error
}

type state int

const (
	inprogressState state = iota
	completedState
	errorState
)

// NewModel creates a new Model.
func NewModel(options ...Option) *Model {
	m := &Model{
		states:          make(map[int]state),
		deletedBranches: make(map[int][]string),
		errMessages:     make(map[int][]error),
		spinner:         newSpinner(),
	}
	for _, option := range options {
		option(m)
	}
	return m
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerColor
	return s
}

// Error returns the error that occurred during the execution of the UI.
func (m *Model) Error() error {
	return m.err
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		// start the ticking of the spinner
		spinner.Tick,
		// load all the repos
		loadRepositories(m.path),
		// handle the first inprocess message
		startProcess(m.startProcessMsgs),
		// handle the first completed message
		completeRepo(m.completedMsgs),
	)
}

func loadRepositories(path string) tea.Cmd {
	return func() tea.Msg {
		repositories, err := getRepositories(path)
		if err != nil {
			return errorMsg{err}
		}
		return repositoriesMsg{repositories}
	}
}

func getRepositories(path string) ([]git.Repository, error) {
	var repositories []git.Repository
	// check if the path given is a repository
	if git.IsGitRepository(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, git.Repository{Path: filepath.Dir(absPath), Name: filepath.Base(absPath)})
	} else {
		// else the path is a directory of containing repositories
		dir, err := os.ReadDir(path)
		if err != nil {
			return repositories, err
		}
		for _, entry := range dir {
			// only consider directories that are git repositories
			if entry.IsDir() && git.IsGitRepository(filepath.Join(path, entry.Name())) {
				repositories = append(repositories, git.Repository{Path: path, Name: entry.Name()})
			}
		}
	}
	return repositories, nil
}

// Update updates the Model and allows the View to be able to be updated.
//
// The message flow is as follows:
// 1. Model.Init is called. This starts the loading process of all repositories and allows the initial inprocess and
//    completed messages to be handled.
// 2. Once the repositoriesMsg is received, the repositories are loaded in the Model and processRepos is called to
//    start processing all repositories.
// 3. Once a inprocessMsg is received, the the Model.inprogress map is updated to allow the view to reelect the process.
//    Then the processing of repositories is started, and handling of the next inprocessMsg is enabled.
// 4. Once a completedMsg is received, the Model is updated based on the start (completed or error), and handling of
//    the next completedMsg is enabled.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle error messages. Immediately quits the programs.
	case errorMsg:
		m.err = msg.err
		return nil, tea.Quit
	// Handle loading repositories. Triggers the processing of the repositories.
	case repositoriesMsg:
		m.repositories = msg.repositories
		return m, m.processRepos()
	// Handle starting the process of a repository. Updates the Model and starts the processing of the specific
	// repository and enables the receiving of the next inprocess message.
	case inprocessMsg:
		m.states[msg.position] = inprogressState
		// use tea.Batch to start multiple commands in parallel
		return m, tea.Batch(
			m.processRepo(msg.position, msg.repository),
			startProcess(m.startProcessMsgs),
		)
	// Handle completing the process of a repository. Updates the model, allows the next repo to be processed and
	// enables receiving of the next completed message.
	case completedMsg:
		if msg.errs != nil {
			m.states[msg.position] = errorState
		} else {
			m.states[msg.position] = completedState
		}
		m.deletedBranches[msg.position] = msg.branches
		m.errMessages[msg.position] = msg.errs
		// allow the next repo to be processed
		m.semaphore.Release(1)
		return m, completeRepo(m.completedMsgs)
	// Handle key presses.
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "j":
			m.viewport.LineUp(1)
		case "down", "k":
			m.viewport.LineDown(1)
		}
		return m, nil
	// Handle spinner ticks.
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	// Handle window resize.
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - 7}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 7
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) processRepos() tea.Cmd {
	return func() tea.Msg {
		for i, r := range m.repositories {
			// limit the number of processes that can process repos
			if err := m.semaphore.Acquire(context.Background(), 1); err != nil {
				return errorMsg{err}
			}
			m.startProcessMsgs <- inprocessMsg{position: i, repository: r}
		}
		return nil
	}
}

func startProcess(startProcessMsgs chan inprocessMsg) tea.Cmd {
	return func() tea.Msg {
		return <-startProcessMsgs
	}
}

func (m *Model) processRepo(position int, repo git.Repository) tea.Cmd {
	return func() tea.Msg {
		go func() {
			branches, errs := process(repo, m.protectedBranches, m.dryRun)
			m.completedMsgs <- completedMsg{position: position, branches: branches, errs: errs}
		}()
		return nil
	}
}

func process(repo git.Repository, protectedBranches []string, dryRun bool) ([]string, []error) {
	fullPath := filepath.Join(repo.Path, repo.Name)
	// Default to "main" branch. If there is an error, will assume the repo's main branch is "master" and try again.
	mainBranch := "main"
	if err := git.CheckoutBranch(fullPath, mainBranch); err != nil {
		mainBranch = "master"
		if err = git.CheckoutBranch(fullPath, mainBranch); err != nil {
			return nil, []error{errors.New("the main branch has not been checked out locally")}
		}
	}
	// ensure everything is up to date so we know for sure which branches are dead (merged)
	if err := git.Pull(fullPath); err != nil {
		return nil, []error{err}
	}
	// get all branches that have been merged into the main branch
	mergedBranches, err := git.GetMergedBranches(fullPath, mainBranch)
	if err != nil {
		return nil, []error{err}
	}
	var branches []string
	var errs []error
	for _, branch := range mergedBranches {
		// skip protected branches
		if !utils.Contains(protectedBranches, branch) {
			// if a dry run, just add the branch to the list of deleted branches
			if dryRun {
				branches = append(branches, branch)
			} else {
				// try to delete the branch
				if err = git.DeleteBranch(fullPath, branch); err != nil {
					errs = append(errs, err)
				} else {
					// if successful, add the branch to the list of deleted branches
					branches = append(branches, branch)
				}
			}
		}
	}
	return branches, errs
}

func completeRepo(completedMsgs chan completedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-completedMsgs
	}
}

func (m *Model) View() string {
	var body string
	if m.ready && len(m.repositories) > 0 {
		m.viewport.SetContent(getBody(m))
		body = m.viewport.View()
	}
	return fmt.Sprintf(
		"%s\n\n%s\n%s",
		getHeader(m),
		body,
		grayStyle.Render(getFooter(m)),
	)
}

func getHeader(m *Model) string {
	if !m.ready || m.repositories == nil {
		return fmt.Sprintf("%s Loading...", m.spinner.View())
	} else if len(m.repositories) == 0 {
		return "There are no repositories in this directory."
	} else {
		completedCount := 0
		errorCount := 0
		for _, s := range m.states {
			if s == completedState {
				completedCount++
			} else if s == errorState {
				errorCount++
			}
		}
		return fmt.Sprintf(
			"%s\n%s",
			fmt.Sprintf("Repositories (%d/%d)", completedCount+errorCount, len(m.repositories)),
			grayStyle.Render(fmt.Sprintf("Branches Deleted - %d", getTotalDeletedBranches(m.deletedBranches))),
		)
	}
}

func getTotalDeletedBranches(deleted map[int][]string) int {
	var total int
	for _, branches := range deleted {
		total += len(branches)
	}
	return total
}

func getBody(m *Model) string {
	defer m.builder.Reset()

	for i, r := range m.repositories {
		if m.states[i] == inprogressState {
			m.builder.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), r.Name))
		} else if m.states[i] == completedState {
			m.builder.WriteString(fmt.Sprintf("%s  %s\n", completedStyle.Render(symbolCheck), r.Name))
		} else if m.states[i] == errorState {
			m.builder.WriteString(fmt.Sprintf("%s  %s\n", errorStyle.Render(symbolX), r.Name))
		} else {
			m.builder.WriteString(fmt.Sprintf("%s  %s\n", " ", r.Name))
		}
		for j, deletedBranch := range m.deletedBranches[i] {
			if j == len(m.deletedBranches[i])-1 {
				m.builder.WriteString(fmt.Sprintf("   %s %s\n", grayStyle.Render(symbolLeaf), grayStyle.Render(deletedBranch)))
			} else {
				m.builder.WriteString(fmt.Sprintf("   %s %s\n", grayStyle.Render(symbolBranch), grayStyle.Render(deletedBranch)))
			}
		}
		for j, err := range m.errMessages[i] {
			if j == len(m.errMessages[i])-1 {
				m.builder.WriteString(fmt.Sprintf("   %s %s\n", errorStyle.Render(symbolLeaf), errorStyle.Render(err.Error())))
			} else {
				m.builder.WriteString(fmt.Sprintf("   %s %s\n", errorStyle.Render(symbolBranch), errorStyle.Render(err.Error())))
			}
		}
	}

	return m.builder.String()
}

func getFooter(m *Model) string {
	return fmt.Sprintf(
		"\n%s\n%s\n%s",
		fmt.Sprintf("Scroll: %3.f%%", m.viewport.ScrollPercent()*100),
		"(press '↑' or '↓' to scroll)",
		"(press 'q' to quit)",
	)
}
