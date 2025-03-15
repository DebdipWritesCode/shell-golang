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

type RedirectionInfo struct {
	outputFile     string
	appendMode     bool
	stdErrRedirect bool
}

func commandIdentifier(command string) {
	formattedCommand := command[:len(command)-1]

	splittedCommands := parseQuotes(formattedCommand)
	firstCommand := splittedCommands[0]

	outputFile, appendMode, stdErrRedirect, splittedCommands := parseRedirect(splittedCommands)

	redirectionInfo := RedirectionInfo{
		outputFile:     outputFile,
		appendMode:     appendMode,
		stdErrRedirect: stdErrRedirect,
	}

	if firstCommand == "exit" {
		handleExit(splittedCommands, redirectionInfo)
		return
	} else if firstCommand == "echo" {
		handleEcho(splittedCommands, redirectionInfo)
		return
	} else if firstCommand == "type" {
		handleType(splittedCommands, redirectionInfo)
		return
	} else if firstCommand == "pwd" {
		handlePwd(redirectionInfo)
		return
	} else if firstCommand == "cd" {
		handleCd(splittedCommands, redirectionInfo)
		return
	} else {
		handleExternalCommands(splittedCommands, redirectionInfo)
		return
	}
}

func parseRedirect(commands []string) (string, bool, bool, []string) {
	var appendMode bool
	var outputFile string
	var stdErrRedirect bool

	for i, arg := range commands {
		if arg == ">" || arg == "1>" || arg == "2>" || arg == ">>" || arg == "1>>" || arg == "2>>" {
			if i+1 < len(commands) {
				outputFile = commands[i+1]
				appendMode = arg == ">>" || arg == "2>>" || arg == "1>>"
				stdErrRedirect = arg == "2>" || arg == "2>>"
				commands = commands[:i]
				break
			}
		}
	}

	return outputFile, appendMode, stdErrRedirect, commands
}

func parseQuotes(command string) []string {
	if !strings.ContainsAny(command, `'"`) {
		return strings.Fields(command)
	}

	var result []string
	var token string
	inSingleQuote, inDoubleQuote := 0, 0

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if ch == '\'' {
			if inDoubleQuote == 1 {
				token += string(ch)
			} else {
				if inSingleQuote == 0 {
					inSingleQuote++
				} else {
					if i+1 < len(command) && command[i+1] == ' ' {
						token += string(" ")
					}
					result = append(result, token)
					token = ""
					inSingleQuote = 0
				}
			}
		} else if ch == '"' {
			if inSingleQuote == 1 {
				token += string(ch)
			} else {
				if inDoubleQuote == 0 {
					inDoubleQuote++
				} else {
					if i+1 < len(command) && command[i+1] == ' ' {
						token += string(" ")
					}
					result = append(result, token)
					token = ""
					inDoubleQuote = 0
				}
			}
		} else if ch == ' ' {
			if inSingleQuote == 1 || inDoubleQuote == 1 {
				token += string(ch)
			} else if token != "" {
				result = append(result, token)
				token = ""
			}
		} else if ch == '\\' {
			if i+1 < len(command) {
				token += string(command[i+1])
				i++
			}
		} else {
			token += string(ch)
		}
	}

	if token != "" {
		result = append(result, token)
	}

	return result
}

func handleExit(commands []string, redirectionInfo RedirectionInfo) {
	if commands[1] != "0" {
		// fmt.Println("exit: " + commands[1] + ": numeric argument required")
		output := "exit: " + commands[1] + ": numeric argument required"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		os.Exit(2)
		return
	}
	os.Exit(0)
	return
}

func handleEcho(commands []string, redirectionInfo RedirectionInfo) {
	output := strings.Join(commands[1:], "")

	handleOutput(output, redirectionInfo.outputFile, redirectionInfo, false)
	return
}

func handleType(commands []string, redirectionInfo RedirectionInfo) {
	knownCommands := []string{
		"echo",
		"exit",
		"type",
		"pwd",
		"cd",
	}

	commandToType := strings.Join(commands, " ")[5:]
	if slices.Contains(knownCommands, commandToType) {
		// fmt.Println(commandToType + " is a shell builtin")
		output := commandToType + " is a shell builtin"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, false)
		return
	} else if path, err := exec.LookPath(commandToType); err == nil {
		// fmt.Println(commandToType + " is " + path)
		output := commandToType + " is " + path
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, false)
		return
	} else {
		// fmt.Println(commandToType + ": not found")
		output := commandToType + ": not found"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		return
	}
}

func handlePwd(redirectionInfo RedirectionInfo) {
	dir, err := os.Getwd()
	if err != nil {
		// fmt.Println("Error getting current directory:", err)
		output := "Error getting current directory: " + err.Error()
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		return
	}
	fmt.Println(dir)
	return
}

func handleCd(commands []string, redirectionInfo RedirectionInfo) {
	if len(commands) > 2 {
		// fmt.Println("cd: too many arguments")
		output := "cd: too many arguments"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
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
		// fmt.Println("cd: " + targetDir + ": No such file or directory")
		output := "cd: " + targetDir + ": No such file or directory"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
	}
}

func handleExternalCommands(commands []string, redirectionInfo RedirectionInfo) {
	_, err := exec.LookPath(commands[0])
	if err != nil {
		output := commands[0] + ": command not found"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		return
	}

	cmd := exec.Command(commands[0], commands[1:]...) // Execute the command with the rest of the arguments

	// Handle redirection if needed
	var outputFile *os.File
	if redirectionInfo.outputFile != "" {
		flag := os.O_CREATE | os.O_WRONLY
		if redirectionInfo.appendMode {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}

		outputFile, err = os.OpenFile(redirectionInfo.outputFile, flag, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer outputFile.Close()

		if redirectionInfo.stdErrRedirect {
			cmd.Stdout = os.Stdout
			cmd.Stderr = outputFile // Redirect stderr
		} else {
			cmd.Stderr = os.Stderr
			cmd.Stdout = outputFile // Redirect stdout
		}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		// output := err.Error()
		// handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		return
	}
}

func handleOutput(output string, outputFile string, redirectionInfo RedirectionInfo, isError bool) {
	if outputFile != "" {
		var file *os.File
		var err error

		if redirectionInfo.appendMode {
			file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			file, err = os.Create(outputFile)
		}

		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		defer file.Close()

		if !isError && !redirectionInfo.stdErrRedirect {
			file.WriteString(output + "\n")
		} else if isError && redirectionInfo.stdErrRedirect {
			file.WriteString(output + "\n")
		} else {
			fmt.Println(output)
		}
	} else {
		if isError {
			fmt.Println(output)
		} else {
			fmt.Println(output)
		}
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
