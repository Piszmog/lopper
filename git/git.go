package git

import (
	"fmt"
	"lopper/utils"
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

// GetMergedSquashedBranches returns a list of merged squashed branches in the given repository.
//
// Credit: https://github.com/not-an-aardvark/git-delete-squashed
func GetMergedSquashedBranches(path string, mainBranch string, mergedBranches []string) ([]string, error) {
	out, err := exec.Command("git", "-C", path, "for-each-ref", "refs/heads/", "--format=%(refname:short)").Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get branches: %s", exitError.Error())
		}
	}
	allBranches := strings.Split(string(out), "\n")
	var squashedBranches []string
	for _, branch := range allBranches {
		if branch == mainBranch {
			continue
		}
		if len(branch) == 0 {
			continue
		}
		// skip merged branches since they were merged commits and will not show up in this process
		if utils.Contains(mergedBranches, branch) {
			continue
		}
		ancestorHash, err := exec.Command("git", "-C", path, "merge-base", mainBranch, branch).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get ancestor hash: %w", err)
		}
		treeId, err := exec.Command("git", "-C", path, "rev-parse", fmt.Sprintf("%s^{tree}", branch)).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get tree id: %w", err)
		}
		danglingCommitId, err := exec.Command("git", "-C", path, "commit-tree", utils.TrimNewline(string(treeId)), "-p", utils.TrimNewline(string(ancestorHash)), "-m", "Temp commit").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get dangling commit id: %w", err)
		}
		commitId, err := exec.Command("git", "-C", path, "cherry", mainBranch, utils.TrimNewline(string(danglingCommitId))).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get commit id: %w", err)
		}
		if strings.HasPrefix(string(commitId), "-") {
			squashedBranches = append(squashedBranches, branch)
		}
	}
	return squashedBranches, nil
}
