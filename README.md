# Custom Shell with Autocomplete and Redirection

## Overview
This project is a custom shell implementation in Golang that supports various built-in commands, redirection features, proper string handling, and an intelligent autocomplete system.

## Features

### 1. Built-in Commands
The shell includes support for the following built-in commands:
- **pwd** - Prints the current working directory.
- **cd [directory]** - Changes the current directory.
- **echo [text]** - Prints the given text, handling quotes and escape sequences correctly.
- **exit** - Exits the shell.

### 2. Input Handling
- Proper handling of **strings, quotes**, and **escape sequences**.
- Allows execution of both built-in and external commands.

### 3. Redirection Support
The shell supports all common output and error redirections:
- **Overwrite mode:**
  - `command > file` (Redirect stdout to a file, overwriting if it exists.)
  - `command 2> file` (Redirect stderr to a file, overwriting if it exists.)
- **Append mode:**
  - `command >> file` (Append stdout to a file.)
  - `command 2>> file` (Append stderr to a file.)
- **Both stdout and stderr:**
  - `command &> file` (Redirect both stdout and stderr to a file.)
  - `command &>> file` (Append both stdout and stderr to a file.)
  
### 4. Autocomplete Feature
The shell implements an **autocomplete system** for command names when the user presses the `Tab` key.
- If multiple commands match the given prefix, it completes to the **longest common prefix**.
- If a single unique match is found, it completes the command automatically.
- If the match is the longest possible valid command, a space is added after it.

Example:
```bash
$ xyz_<TAB>        # Completes to xyz_foo
$ xyz_foo_<TAB>    # Completes to xyz_foo_bar
$ xyz_foo_bar_<TAB> # Completes to xyz_foo_bar_baz (with a space at the end)
```

### 5. Executable Handling
- The shell searches for executables in the system `PATH`.
- Supports both **built-in commands** and **external programs**.
- Implements `longestCommonPrefix` logic for intelligent tab completion.

## Installation and Usage
### 1. Clone the Repository
```sh
git clone https://github.com/DebdipWritesCode/shell-golang.git
cd shell-golang
cd app
```

### 2. Build the Shell
```sh
go build -o my_shell main.go
```

### 3. Run the Shell
```sh
./my_shell
```

## Future Enhancements
- Implement **piping (`|`)** between commands.
- Support for **background processes (`&`)**.
- Enhanced **command history and navigation**.

## Contributors
- **Debdip Mukherjee** (@DebdipWritesCode)

## License
This project is licensed under the MIT License.