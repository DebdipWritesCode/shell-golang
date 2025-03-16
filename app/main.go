package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

type RedirectionInfo struct {
	outputFile     string
	appendMode     bool
	stdErrRedirect bool
}

var builtInCommands = []string{
	"exit",
	"echo",
	"type",
	"pwd",
	"cd",
}

func commandIdentifier(command string) {
	splittedCommands := parseQuotes(command)
	firstCommand := splittedCommands[0]

	firstCommand = strings.TrimSpace(firstCommand)

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
		arg = strings.TrimSpace(arg)
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
					if i+1 < len(command) {
						if command[i+1] == '\'' {
							i++
							continue
						} else if command[i+1] == '"' {
							inDoubleQuote++
							inSingleQuote = 0
							i++
							continue
						} else if command[i+1] == ' ' {
							token += string(" ")
						}
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
					if i+1 < len(command) {
						if command[i+1] == '"' {
							i++
							continue
						} else if command[i+1] == '\'' {
							inDoubleQuote = 0
							inSingleQuote++
							i++
							continue
						} else if command[i+1] == ' ' {
							token += string(" ")
						}
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
				token += string(ch)
				result = append(result, token)
				token = ""
			}
		} else if ch == '\\' {
			if inSingleQuote == 1 {
				token += string(ch)
				continue
			} else if inDoubleQuote == 1 {
				if i+1 < len(command) {
					if command[i+1] == '\\' || command[i+1] == '"' || command[i+1] == '$' {
						token += string(command[i+1])
						i++
						continue
					} else {
						token += string(ch)
					}
				} else {
					token += string(ch)
				}
			}
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

func trimCommands(commands []string) {
	for i, command := range commands {
		commands[i] = strings.TrimSpace(command)
	}
}

func handleExit(commands []string, redirectionInfo RedirectionInfo) {
	trimCommands(commands)
	if commands[1] != "0" {
		// fmt.Println("exit: " + commands[1] + ": numeric argument required")
		output := "exit: " + commands[1] + ": numeric argument required"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		os.Exit(2)
		return
	}
	os.Exit(0)
}

func handleEcho(commands []string, redirectionInfo RedirectionInfo) {
	output := strings.Join(commands[1:], "")

	handleOutput(output, redirectionInfo.outputFile, redirectionInfo, false)
}

func handleType(commands []string, redirectionInfo RedirectionInfo) {
	knownCommands := []string{
		"echo",
		"exit",
		"type",
		"pwd",
		"cd",
	}
	trimCommands(commands)

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
}

func handleCd(commands []string, redirectionInfo RedirectionInfo) {
	trimCommands(commands)
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
	trimCommands(commands)
	_, err := exec.LookPath(commands[0])
	if err != nil {
		output := commands[0] + ": command not found"
		handleOutput(output, redirectionInfo.outputFile, redirectionInfo, true)
		return
	}

	for i, command := range commands {
		if command == ">" || command == "1>" || command == "2>" || command == ">>" || command == "1>>" || command == "2>>" {
			commands = commands[:i]
			break
		}
	}
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
			handleOutput(err.Error(), redirectionInfo.outputFile, redirectionInfo, true)
			return
		}
		defer outputFile.Close()
	}

	cmd := exec.Command(commands[0], commands[1:]...) // Execute the command with the rest of the arguments

	if redirectionInfo.outputFile != "" {
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

func getExecutablesFromPath(prefix string) []string {
	matches := make(map[string]bool) // Use a map to avoid duplicates

	pathDirs := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))

	for _, dir := range pathDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
				matches[file.Name()] = true
			}
		}
	}

	var uniqueMatches []string
	for match := range matches {
		uniqueMatches = append(uniqueMatches, match)
	}

	return uniqueMatches
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	fmt.Println(strs)

	prefix := strs[0]
	for _, str := range strs[1:] {
		for strings.Index(str, prefix) != 0 {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}

	if strs[len(strs)-1] == prefix {
		return prefix + " "
	}

	return prefix
}

func autoComplete(line string) []string {
	suggestions := make(map[string]bool)

	for _, cmd := range getExecutablesFromPath(line) {
		suggestions[cmd] = true
	}

	for _, cmd := range builtInCommands {
		if strings.HasPrefix(cmd, line) {
			suggestions[cmd] = true
		}
	}

	var uniqueSuggestions []string
	for suggestion := range suggestions {
		uniqueSuggestions = append(uniqueSuggestions, suggestion)
	}

	sort.Strings(uniqueSuggestions)

	if len(suggestions) > 1 {
		commonPrefix := longestCommonPrefix(uniqueSuggestions)
		if commonPrefix != line { // Auto-complete only if LCP is longer than current input
			return []string{commonPrefix}
		}
	}

	return uniqueSuggestions
}

func getTermios(fd int) (*syscall.Termios, error) {
	var t syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return &t, nil
}

// Set terminal attributes
func setTermios(fd int, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(t)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func main() {
	// Uncomment this block to pass the first stage

	// Wait for user input
	// bufio.NewReader(os.Stdin).ReadString('\n')
	// status := 0

	for {
		fmt.Fprint(os.Stdout, "$ ")

		fd := int(os.Stdin.Fd()) // Get the file descriptor for the standard input

		originalTermios, err := getTermios(fd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting terminal attributes:", err)
			os.Exit(1)
		}
		defer setTermios(fd, originalTermios)

		raw := *originalTermios
		raw.Lflag &^= syscall.ECHO | syscall.ICANON // Disable echo and canonical mode
		raw.Cc[syscall.VMIN] = 1                    // Minimum number of characters to read
		raw.Cc[syscall.VTIME] = 0                   // No timeout
		if err := setTermios(fd, &raw); err != nil {
			fmt.Println("Error setting terminal to raw mode:", err)
			return
		}

		buffer := make([]byte, 1)       // Buffer to read a single byte
		input_buffer := make([]byte, 0) // Buffer to store the input
		splt_input := []string{}
		tab_pressed := false
		for {
			_, err := os.Stdin.Read(buffer)
			if err != nil {
				fmt.Println("Error reading input: ", err)
				break
			}

			if buffer[0] == '\t' {
				if len(input_buffer) == 0 {
					continue
				}
				splt_input = strings.Split(string(input_buffer), " ")
				suggestions := autoComplete(splt_input[0])

				if len(suggestions) == 0 {
					tab_pressed = false
					fmt.Print("\a") // Beep sound suggesting no suggestions
				} else if len(suggestions) == 1 {
					tab_pressed = false

					toOutput := suggestions[0]
					for _, cmd := range builtInCommands {
						if cmd == toOutput {
							toOutput = toOutput + " "
							break
						}
					}

					input_buffer = []byte(toOutput) // Auto-complete with the only match
					fmt.Print("\r\x1b[K")           // Clears the line
					fmt.Printf("$ %s", input_buffer)
				} else if len(suggestions) > 1 {
					if tab_pressed {
						fmt.Println()
						for _, suggestion := range suggestions {
							fmt.Printf("%s  ", suggestion)
						}
						fmt.Println()
						fmt.Printf("$ %s", input_buffer) // KEEP the original input buffer
						tab_pressed = false
					} else {
						tab_pressed = true
						fmt.Print("\a") // Beep on first tab press
					}
				}
			} else if buffer[0] == '\n' {
				fmt.Print("\n")
				splt_input = strings.Split(string(input_buffer), " ")
				commandIdentifier(string(input_buffer))
				break
			} else if buffer[0] == '\x7f' { // Backspace
				tab_pressed = false
				if len(input_buffer) > 0 {
					fmt.Print("\r\x1b[K") // This clears the line
					input_buffer = input_buffer[:len(input_buffer)-1]
					fmt.Printf("$ %s", input_buffer)
				} else if tab_pressed {
					// To be done
				}
			} else {
				tab_pressed = false
				input_buffer = append(input_buffer, buffer[0])
				fmt.Printf("%s", string(buffer[0]))
			}
		}
	}

	// for {
	// 	fmt.Fprint(os.Stdout, "$ ")
	// 	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	// 		os.Exit(1)
	// 	}

	// 	commandIdentifier(command)
	// }
}
