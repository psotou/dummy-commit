package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func GitCommand(args ...string) (*exec.Cmd, error) {
	// for the time being, I'm only focusing on linux machines. For a safer alternative
	// that includes the windows runtime, use: github.com/cli/safeexec
	gitExe, err := exec.LookPath("git")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errors.New("unable to find git executable in PATH, please install git before retrying")
		}
		return nil, err
	}
	return exec.Command(gitExe, args...), nil
}

type Commit struct {
	Sha   string
	Title string
}

// Commits returns the commits between a base ref branch and the head ref branch
func Commits(baseRef, headRef string) ([]*Commit, error) {
	logCmd, err := GitCommand(
		"-c", "log.ShowSignature=false",
		"log", "--pretty=format:%H,%s",
		"--cherry", fmt.Sprintf("%s...%s", baseRef, headRef))
	if err != nil {
		return nil, err
	}
	output, err := logCmd.Output()
	if err != nil {
		return []*Commit{}, err
	}

	var commits []*Commit
	sha := 0
	title := 1
	for _, line := range outputLines(output) {
		split := strings.SplitN(line, ",", 2)
		if len(split) != 2 {
			continue
		}
		commits = append(commits, &Commit{
			Sha:   split[sha],
			Title: split[title],
		})
	}

	if len(commits) == 0 {
		return commits, fmt.Errorf("could not find any commits between %s and %s", baseRef, headRef)
	}

	return commits, nil
}

func dummyCommitSha(commits []*Commit) string {
	// think about this message, dummy commit may be too generic
	message := "dummy commit"
	for _, commit := range commits {
		if strings.Contains(commit.Title, message) {
			return commit.Sha
		}
	}
	return ""
}

func numberOfCommits(commits []*Commit) string {
	// we need to add 1 to the number of commits present in the current branch
	// because we use this helper function when we do a fixup, and therefore add a new commit
	// with fixup! message
	return strconv.Itoa(len(commits) + 1)
}

// CurrentBranch reads the checked-out branch for the git repository
func currentBranch() (string, error) {
	refCmd, err := GitCommand("symbolic-ref", "--quiet", "HEAD")
	if err != nil {
		return "", err
	}

	stdErr := bytes.Buffer{}
	refCmd.Stderr = &stdErr

	output, err := refCmd.Output()
	if err == nil {
		return getBranchShortName(output), nil
	}
	return "", fmt.Errorf("%s git: %s", stdErr.String(), err)
}

// from here on starts the flow to add, fixup and squash the commit
func GitAdd(file string) error {
	gitCmd, err := GitCommand("add", file)
	if err != nil {
		return err
	}
	log.Printf("git: adding %s file", readmeFile)
	return gitCmd.Run()
}

func GitCommit(message string) error {
	gitCmd, err := GitCommand("commit", "-m", message)
	if err != nil {
		return err
	}
	log.Println("git: commiting file")
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

func GitFixup(commitSha string) error {
	gitCmd, err := GitCommand("commit", "--fixup", commitSha)
	if err != nil {
		return err
	}
	log.Printf("git: running commit --fixup on commit %s", commitSha)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

// this opens up an editot, if you have VIM configured by default save and quit to successfully rebase your fixups
func GitRebase(numberOfCommits string) error {
	gitCmd, err := GitCommand("rebase", "--interactive", "--autosquash", "HEAD~"+numberOfCommits)
	if err != nil {
		return err
	}
	log.Printf("git: running rebase -i --autosquash on %s commits", numberOfCommits)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

// TODO
// for the time being, pass the remote manually since in the project we are working is fixed
// to be origin. But do remember to work on this since it should not be hardcoded :s

// func Push(remote string, ref string, cmdOut, cmdErr io.Writer) error {
func GitPush(remote string, ref string) error {
	gitCmd, err := GitCommand("push", "--set-upstream", remote, ref)
	if err != nil {
		return err
	}
	log.Printf("git: pushing to %s %s", remote, ref)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

func GitPushForce(remote string, ref string) error {
	gitCmd, err := GitCommand("push", "--set-upstream", "--force-with-lease", remote, ref)
	if err != nil {
		return err
	}
	log.Printf("git: pushing to %s %s", remote, ref)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

/*
This might not be necessary, since you're already storing the commit title in the Commit struct.
So I believe a simple strings contains should do the job iterating over the commit titles stored in []Commit
*/
// func lookupCommit(sha, format string) ([]byte, error) {
// 	logCmd, err := GitCommand("-c", "log.ShowSignature=false", "show", "-s", "--pretty=format:"+format, sha)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return logCmd.Output()
// }

func outputLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")

}

func firstLine(output []byte) string {
	if i := bytes.IndexAny(output, "\n"); i >= 0 {
		return string(output)[0:i]
	}
	return string(output)
}

func getBranchShortName(output []byte) string {
	branch := firstLine(output)
	return strings.TrimPrefix(branch, "refs/heads/")
}
