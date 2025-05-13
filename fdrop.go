package main

import (

	"encoding/json"
	"fmt"
	"io"

	"os"
	"path/filepath"
	"strconv"

	"time"
)
const (
	version = "2.0.0"
)


type StashItem struct {
	Path string
}

var (
	homeDir, _       = os.UserHomeDir()
	stashDir         = filepath.Join(homeDir, ".fdrop")
	stashFilePath    = filepath.Join(stashDir, "stash.json")
	logFilePath      = filepath.Join(stashDir, "log.txt")
)

func ensureDirs() {
	os.MkdirAll(stashDir, 0755)
}

func writeLog(action, src, dest string) {
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		logLine := fmt.Sprintf("%s | %s | %s ➜ %s\n", time.Now().Format(time.RFC3339), action, src, dest)
		f.WriteString(logLine)
	}
}

func readStash() ([]StashItem, error) {
	data, err := os.ReadFile(stashFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []StashItem{}, nil
		}
		return nil, err
	}
	var stash []StashItem
	err = json.Unmarshal(data, &stash)
	return stash, err
}

func writeStash(stash []StashItem) error {
	data, err := json.MarshalIndent(stash, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stashFilePath, data, 0644)
}

func addToStash(paths []string) {
	ensureDirs()
	stash, _ := readStash()
	seen := map[string]bool{}

	for _, item := range stash {
		seen[item.Path] = true
	}

	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil || abs == "." {
			fmt.Printf("Invalid path: %s\n", p)
			continue
		}
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", p)
			continue
		}
		if seen[abs] {
			fmt.Printf("File already in stash from this directory: %s\n", abs)
			continue
		}

		// Check if same base name exists
		for _, existing := range stash {
			if filepath.Base(existing.Path) == filepath.Base(abs) && existing.Path != abs {
				fmt.Printf("Warning: Duplicate file name from different directory: %s (existing: %s)\n", abs, existing.Path)
			}
		}

		stash = append(stash, StashItem{Path: abs})
		fmt.Printf("Added: %s from %s\n", filepath.Base(abs), filepath.Dir(abs))
	}
	writeStash(stash)
}

func listStash() {
	stash, _ := readStash()
	for i, item := range stash {
		fmt.Printf("%d. %s (from %s)\n", i+1, filepath.Base(item.Path), filepath.Dir(item.Path))
	}
}

func resolveItems(args []string, stash []StashItem) ([]StashItem, []int) {
	var result []StashItem
	var indexes []int
	for _, arg := range args {
		i, err := strconv.Atoi(arg)
		if err == nil && i > 0 && i <= len(stash) {
			result = append(result, stash[i-1])
			indexes = append(indexes, i-1)
		} else {
			found := false
			for idx, item := range stash {
				if filepath.Base(item.Path) == arg {
					result = append(result, item)
					indexes = append(indexes, idx)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("not found in stash: %s\n", arg)
			}
		}
	}
	return result, indexes
}

func copyFile(src, dest string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	info, err := srcF.Stat()
	if err != nil {
		return err
	}

	destF, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destF.Close()

	_, err = io.Copy(destF, srcF)
	return err
}

func performAction(action string, args []string, keep bool) {
	ensureDirs()
	stash, _ := readStash()
	target := "."
	files := args

	if len(args) > 1 {
		if fi, err := os.Stat(args[len(args)-1]); err == nil && fi.IsDir() {
			target = args[len(args)-1]
			files = args[:len(args)-1]
		}
	}

	items, indexes := resolveItems(files, stash)
	var newStash []StashItem
	for i, item := range stash {
		shouldSkip := false
		for _, idx := range indexes {
			if idx == i {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip || keep {
			newStash = append(newStash, item)
		}
	}

	for _, item := range items {
		destPath := filepath.Join(target, filepath.Base(item.Path))
		if action == "move" {
			err := os.Rename(item.Path, destPath)
			if err != nil {
				fmt.Printf("Error moving %s: %v\n", item.Path, err)
				continue
			}
		} else {
			err := copyFile(item.Path, destPath)
			if err != nil {
				fmt.Printf("Error copying %s: %v\n", item.Path, err)
				continue
			}
		}
		fmt.Printf("%s ➜ %s\n", filepath.Base(item.Path), filepath.Base(destPath))
		writeLog(action, item.Path, destPath)
	}
	writeStash(newStash)
}

func cleanStash() {
	_ = writeStash([]StashItem{})
	fmt.Println("Stash cleaned.")
}

func showLogs() {
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		fmt.Println("No logs found.")
		return
	}
	fmt.Println(string(data))
}
func showVersion() {
	fmt.Printf("fdrop version %s\n", version)
}

func printHelp() {
	fmt.Println(`fdrop - A clipboard-like file copy-paste tool for the terminal.

Usage:
  fdrop add <file|dir>...           Add files/directories to the stash.
  fdrop copy <file|index> [dir]     Copy specific file(s) to target directory.
  fdrop move <file|index> [dir]     Move specific file(s) to target directory.
  fdrop paste [target_dir]          Copy all stashed files to directory.
  fdrop stash                       List all stashed files.
  fdrop stash keep <file|index> [dir]  Copy but retain file in stash.
  fdrop clean                       Clear the stash.
  fdrop --logs                      Show action logs.
  fdrop --version                   Show the version of fdrop.
  fdrop --help                      Show this help.

  Support:
  Email: akhileshs2220@gmail.com
`)
}

func main() {
	ensureDirs()
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "add":
		addToStash(args[1:])
	case "stash":
		if len(args) == 1 {
			listStash()
		} else if args[1] == "keep" {
			performAction("copy", args[2:], true)
		} else {
			fmt.Println("Invalid command. Usage: fdrop stash keep <filename|index> [target_dir]")
		}
	case "copy":
		performAction("copy", args[1:], false)
	case "move":
		performAction("move", args[1:], false)
	case "paste":
		stash, _ := readStash()
		names := []string{}
		for i := range stash {
			names = append(names, strconv.Itoa(i+1))
		}
		if len(args) > 1 {
			names = args[1:]
		}
		performAction("copy", names, false)
	case "clean":
		cleanStash()
	case "--logs":
		showLogs()
	case "--help":
		printHelp()
	case "--version":
		showVersion()
	default:
		fmt.Println("Invalid command. Use fdrop --help.")
	}
}
