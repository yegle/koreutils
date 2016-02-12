package cmd

import (
	"log"
	"os"
)

var stderr = log.New(os.Stderr, "", 0)

type ExitMessage struct {
	Message string
}

func (e ExitMessage) Error() string {
	return e.Message
}

func ExitError(err error) {
	stderr.Println(err)
	os.Exit(1)
}
