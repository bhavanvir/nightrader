package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
	"runtime"
	"time"
)

func main() {
    args := os.Args

    if len(args) == 1 {
        // If no argument provided, execute all commands
        executeAllCommands()
    } else if len(args) == 2 {
        // If argument is "test", execute commands 1 to 3
        executeCommands1to3()
    } else {
        fmt.Println("Invalid arguments. Usage: go run ./main.go [test]")
    }
}

func executeAllCommands() {
    // 1. Docker compose down database
    executeCommand("docker-compose", "down", "database")

    // 2. Docker compose up -d database
    executeCommand("docker-compose", "up", "-d", "database")

    // 3. Delete a file if exists: tests\TestRun2\Config\stockIds.csv
    deleteFileIfExists("tests/TestRun2/Config/stockIds.csv")
    deleteFileIfExists("tests/TestRun3/Config/stockIds.csv")

    // 4. Open new terminal: cd .\engine\. Then go run .\main.go
    executeCommandInNewTerminal("engine", "go", "run", "./main.go")

    // 5. Open new terminal: cd .\authentication\. Then go run .\main.go
    executeCommandInNewTerminal("authentication", "go", "run", "./main.go")

    // 6. Open new terminal: cd .\setup\. Then go run .\main.go
    executeCommandInNewTerminal("setup", "go", "run", "./main.go")

    // 7. Open new terminal: cd .\transaction\. Then go run .\main.go
    executeCommandInNewTerminal("transaction", "go", "run", "./main.go")

    // 8. Open new terminal: cd .\execution\. Then go run .\main.go
    executeCommandInNewTerminal("execution", "go", "run", "./main.go")
}

func executeCommands1to3() {
    // 1. Docker compose down database
    executeCommand("docker-compose", "down", "database")

    // 2. Docker compose up -d database
    if err := executeCommand("docker-compose", "up", "-d", "database"); err != nil {
        fmt.Println("Error starting the database:", err)
        os.Exit(1)
    }

    // 3. Delete a file if exists: tests\TestRun2\Config\stockIds.csv
    deleteFileIfExists("tests/TestRun2/Config/stockIds.csv")
    deleteFileIfExists("tests/TestRun3/Config/stockIds.csv")

    // Wait for 10 seconds
    time.Sleep(7 * time.Second)
}

func executeCommand(command string, args ...string) error {
    cmd := exec.Command(command, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func deleteFileIfExists(filePath string) {
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        fmt.Println("Error getting absolute path:", err)
        os.Exit(1)
    }

    if _, err := os.Stat(absPath); !os.IsNotExist(err) {
        err := os.Remove(absPath)
        if err != nil {
            fmt.Println("Error deleting file:", err)
            os.Exit(1)
        }
        fmt.Println("File", absPath, "deleted successfully")
    }
}

func executeCommandInNewTerminal(directory string, command string, args ...string) {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
		cmd = exec.Command("cmd", "/C", "start", "cmd", "/k", "cd .\\"+directory+"\\ && go run .\\main.go")
    case "linux", "darwin": // Unix-like systems
		cmd = exec.Command("gnome-terminal", "--", "bash", "-c", "cd ./"+directory+"; go run main.go")
    default:
        fmt.Println("Unsupported operating system")
        return
    }

    if err := cmd.Run(); err != nil {
        fmt.Println("Error executing command in new terminal:", err)
        os.Exit(1)
    }
}
