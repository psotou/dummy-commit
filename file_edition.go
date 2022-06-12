package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

const readmeFile = "README.md"

var (
	dummyComment    = []byte("\n<!-- dummy commit: ")
	dummyCommentOn  = []byte("\n<!-- dummy commit: on -->\n")
	dummyCommentOff = []byte("\n<!-- dummy commit: off -->\n")
)

var (
	errorInstanceNotFound = fmt.Errorf("no instance of %s was found", dummyComment)
)

func toggleCommentInFile() (int, error) {
	file, err := os.OpenFile(readmeFile, os.O_RDWR, 0644)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	commentIndex := dummyCommentIndex()
	if commentIndex == -1 {
		return file.WriteAt(dummyCommentOn, commentIndex)
	}
	if isCommentOn() {
		return file.WriteAt(dummyCommentOff, commentIndex)
	}
	return file.WriteAt(dummyCommentOn, commentIndex)
}

func isCommentOn() bool {
	data, err := os.ReadFile(readmeFile)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.Contains(data, dummyCommentOn)
}

func dummyCommentIndex() int64 {
	data, err := os.ReadFile(readmeFile)
	if err != nil {
		log.Fatal(err)
	}
	if bytes.Index(data, dummyComment) == -1 {
		return int64(len(data) - 1)
	}

	return int64(bytes.Index(data, dummyComment))
}
