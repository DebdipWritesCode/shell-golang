package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func commandIdentifier(command string) {
	formattedCommand := command[:len(command)-1]

	splittedCommands := strings.Split(formattedCommand, " ")
	firstCommand := splittedCommands[0]

	var appendMode bool
	var outputFile string
	var stdErrRedirect bool

	for i, arg := range splittedCommands {
		if arg == ">" || arg == "1>" || arg == "2>" || arg == ">>" || arg == "2>>" {
			if i+1 < len(splittedCommands) {
				outputFile = splittedCommands[i+1]
				appendMode = arg == ">>" || arg == "2>>"
				stdErrRedirect = arg == "2>" || arg == "2>>"
				splittedCommands = splittedCommands[:i]
				break
			}
		}
	}

	if firstCommand == "exit" {
		handleExit(splittedCommands)
		return
	} else if firstCommand == "echo" {
		handleEcho(splittedCommands)
		return
	} else if firstCommand == "type" {
		handleType(splittedCommands)
		return
	} else if firstCommand == "pwd" {
		handlePwd()
		return
	} else if firstCommand == "cd" {
		handleCd(splittedCommands)
		return
	} else {
		handleExternalCommands(splittedCommands, outputFile, appendMode, stdErrRedirect)
		return
	}
}

func handleExit(commands []string) {
	if commands[1] != "0" {
		fmt.Println("exit: " + commands[1] + ": numeric argument required")
		os.Exit(2)
		return
	}
	os.Exit(0)
	return
}

func handleEcho(commands []string) {
	totalToPrint := strings.Join(commands, " ")[5:]
	fmt.Println(totalToPrint)
	return
}

func handleType(commands []string) {
	knownCommands := []string{
		"echo",
		"exit",
		"type",
		"pwd",
		"cd",
	}

	commandToType := strings.Join(commands, " ")[5:]
	if slices.Contains(knownCommands, commandToType) {
		fmt.Println(commandToType + " is a shell builtin")
		return
	} else if path, err := exec.LookPath(commandToType); err == nil {
		fmt.Println(commandToType + " is " + path)
		return
	} else {
		fmt.Println(commandToType + ": not found")
		return
	}
}

func handlePwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	fmt.Println(dir)
	return
}

func handleCd(commands []string) {
	if len(commands) > 2 {
		fmt.Println("cd: too many arguments")
		return
	}

	targetDir := ""
	if len(commands) == 1 || commands[1] == "~" {
		targetDir, _ = os.UserHomeDir()
	} else {
		targetDir = commands[1]
	}

	err := os.Chdir(targetDir)
	if err != nil {
		fmt.Println("cd: " + targetDir + ": No such file or directory")
	}
}

func handleExternalCommands(commands []string, outputFile string, appendMode bool, stdErrRedirect bool) {
	_, err := exec.LookPath(commands[0])
	if err != nil {
		fmt.Println(commands[0] + ": command not found")
		return
	}

	cmd := exec.Command(commands[0], commands[1:]...) // Execute the command with the rest of the arguments
	if outputFile != "" {
		var file *os.File
		var err error

		if appendMode {
			file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Open the file in append mode if it exists or create it if it doesn't and write to it and 0644 is the permission for the file
		} else {
			file, err = os.Create(outputFile) // Create the file if it doesn't exist and write to it
		}

		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		defer file.Close() // Close the file after the function ends

		if stdErrRedirect {
			cmd.Stderr = file
		} else {
			cmd.Stdout = file
		}

	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing", commands[0], ":", err)
		return
	}
}

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

		commandIdentifier(command)
	}
}
