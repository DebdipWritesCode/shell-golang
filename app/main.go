package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	// Uncomment this block to pass the first stage

	// Wait for user input
	// bufio.NewReader(os.Stdin).ReadString('\n')
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

		if strings.Split(formattedCommand, " ")[0] == "echo" {
			fmt.Println(formattedCommand[5:])
			continue
		}

		fmt.Println(formattedCommand + ": command not found")
	}
}
