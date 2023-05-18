package main

import (
	"fmt"
	"os"

	"github.com/rneatherway/gh-slack/cmd/gh-slack/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
