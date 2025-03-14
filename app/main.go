package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	// Uncomment this block to pass the first stage

	// Wait for user input
	// bufio.NewReader(os.Stdin).ReadString('\n')
	knownCommands := []string{
		"echo",
		"exit",
		"type",
	}

	for {
		fmt.Fprint(os.Stdout, "$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		formattedCommand := command[:len(command)-1]

		if formattedCommand == "exit 0" {
			os.Exit(0)
		}

		firstCommand := strings.Split(formattedCommand, " ")[0]

		if firstCommand == "echo" {
			fmt.Println(formattedCommand[5:])
			continue
		} else if firstCommand == "type" {
			commandToType := formattedCommand[5:]
			if slices.Contains(knownCommands, commandToType) {
				fmt.Println(commandToType + " is a shell builtin")
				continue
			} else {
				fmt.Println(commandToType + " not found")
			}
		}

		fmt.Println(formattedCommand + ": command not found")
	}
}
