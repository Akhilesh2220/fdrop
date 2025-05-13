package main

import (
	"fmt"
	"os"
	"time"
	"io"
	"path/filepath"
)

const clipboardFile = "/tmp/fdrop_clipboard"
const logFile = "/tmp/fdrop_log"
const version = "1.0.0" // Update with your current version

func printHelp() {
	fmt.Println(`fdrop - A clipboard-like file copy-paste tool for the terminal.

Usage:
  fdrop cp <file|dir>       Copy a file or directory to the clipboard.
  fdrop paste [target_dir]   Paste the copied file into the target directory.

Options:
  --help    Show this help message.
  --version Show the version of fdrop.
`)
}

func printVersion() {
	fmt.Printf("fdrop version %s\n", version)
}

func logAction(action string) {
	// Open the log file in append mode, creating it if necessary
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	// Write the log action with a newline at the end
	file.WriteString(fmt.Sprintf("%s\n%s\n\n", time.Now().Format(time.RFC1123), action))
}

func cp(filePath string) {
	// Get the absolute file path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	// Save the path to clipboard file
	err = os.WriteFile(clipboardFile, []byte(absPath), 0644)
	if err != nil {
		fmt.Println("Error writing to clipboard file:", err)
		return
	}

	// Log the action
	logAction(fmt.Sprintf("Copied: %s", absPath))

	fmt.Printf("Copied: %s\n", absPath)
}

func paste(targetDir string) {
	// Read the clipboard file to get the copied file path
	data, err := os.ReadFile(clipboardFile)
	if err != nil {
		fmt.Println("Error reading clipboard file:", err)
		return
	}
	absPath := string(data)

	// Check if target directory is provided, if not, use current directory
	if targetDir == "" {
		targetDir = "."
	}

	// Get the base file name
	baseName := filepath.Base(absPath)

	// Define the target path
	targetPath := filepath.Join(targetDir, baseName)

	// Copy or move the file
	err = copyFile(absPath, targetPath)
	if err != nil {
		fmt.Println("Error pasting file:", err)
		return
	}

	// Log the action
	logAction(fmt.Sprintf("Pasted: %s to %s", baseName, targetDir))

	fmt.Printf("Pasted: %s to %s\n", baseName, targetDir)
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the content from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Set the destination file permissions to match the source
	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	err = os.Chmod(dst, sourceFileInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set permissions on destination file: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	switch command {
	case "cp":
		if len(os.Args) < 3 {
			fmt.Println("Usage: fdrop cp <file|dir>")
			return
		}
		cp(os.Args[2])
	case "paste":
		targetDir := ""
		if len(os.Args) > 2 {
			targetDir = os.Args[2]
		}
		paste(targetDir)
	case "--help":
		printHelp()
	case "--version":
		printVersion()
	default:
		fmt.Println("Invalid command. Usage: fdrop cp <file|dir> or fdrop paste [target_dir]")
	}
}

