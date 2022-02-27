package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Repository represents a Git repository.
type Repository struct {
	Path string
	Name string
}

// IsGitRepository returns true if the given path is a Git repository.
func IsGitRepository(path string) bool {
	if err := exec.Command("git", "-C", path, "rev-parse").Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false
		}
	}
	return true
}

// CheckoutBranch checks out the given branch in the given repository.
func CheckoutBranch(path string, branch string) error {
	if err := exec.Command("git", "-C", path, "checkout", branch).Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("failed to checkout branch %s", branch)
		}
	}
	return nil
}

// Pull updates the given repository.
func Pull(path string) error {
	if err := exec.Command("git", "-C", path, "pull").Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 1:
				return fmt.Errorf("remote repository not found")
			case 128:
				return fmt.Errorf("there is a conflict between remote and local changes")
			default:
				return fmt.Errorf("failed to pull latest changes: %s", exitError.Error())
			}
		}
	}
	return nil
}

// DeleteBranch deletes the given branch in the given repository.
func DeleteBranch(path string, branch string) error {
	if err := exec.Command("git", "-C", path, "branch", "-d", branch).Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("failed to delete branch %s: %s", branch, exitError.Error())
		}
	}
	return nil
}

var branchReplacer = strings.NewReplacer("*", "", " ", "")

// GetMergedBranches returns a list of merged branches in the given repository.
func GetMergedBranches(path string, mainBranch string) ([]string, error) {
	out, err := exec.Command("git", "-C", path, "branch", "--merged", mainBranch).Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get merged branches: %s", exitError.Error())
		}
	}
	allBranches := strings.Split(string(out), "\n")
	var mergedBranches []string
	for _, branch := range allBranches {
		b := branchReplacer.Replace(branch)
		if len(b) > 0 && b != mainBranch {
			mergedBranches = append(mergedBranches, b)
		}
	}
	return mergedBranches, nil
}
