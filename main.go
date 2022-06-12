package main

import (
	"fmt"
	"log"
)

func main() {
	// modify README.md file
	bytesWritten, err := toggleCommentInFile()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("file: %d bytes written to %s", bytesWritten, readmeFile)

	currentbranch, err := currentBranch()
	if err != nil {
		fmt.Println(err)
	}
	// we look for the commits present in the current branch
	commits, err := Commits("main", currentbranch)
	if err != nil {
		fmt.Println(err)
	}
	numberOfCommits := numberOfCommits(commits)

	dummySha := dummyCommitSha(commits)
	if dummySha == "" {
		log.Println("no dummy commit present in the current branch")
		log.Println("adding dummy commit to current branch")
		GitAdd(readmeFile)
		GitCommit("dummy commit")
		GitPush("origin", currentbranch)
	}

	if dummySha != "" {
		log.Printf("dummy commit found %s", dummySha)
	}
	GitAdd(readmeFile)
	GitFixup(dummySha)
	GitRebase(numberOfCommits)
	GitPushForce("origin", currentbranch)
}
