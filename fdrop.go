package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	stashFile = "/tmp/fdrop_stash"
	logFile   = "/tmp/fdrop_log"
	version   = "2.1.0"
)

type StashItem struct {
	Name string
	Path string
}

func main() {
	args := os.Args
	if len(args) < 2 {
		printHelp()
		return
	}

	switch args[1] {
	case "--help":
		printHelp()
	case "--version":
		fmt.Println("fdrop version", version)
	case "--logs":
		showLogs()
	case "stash":
		showStash()
	case "add":
		addToStash(args[2:])
	case "copy":
		pasteFiles(args[2:], false)
	case "move":
		pasteFiles(args[2:], true)
	case "paste":
		pasteFiles([]string{}, false)
	default:
		fmt.Println("Unknown command:", args[1])
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`fdrop - Clipboard-like file tool

Commands:
  fdrop add <file1> <file2> ...        Add file(s)/folder(s) to stash
  fdrop copy <name|stash idx> [...]    Copy from stash to target or .
  fdrop move <name|stash idx> [...]    Move from stash
  fdrop paste                          Paste everything from stash
  fdrop stash                          Show stashed files
  fdrop --logs                         View action log
  fdrop --help                         Show help
  fdrop --version                      Show version`)
}

func logAction(entry string) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] %s\n", ts, entry)
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(msg)
}

func showLogs() {
	data, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Println("Failed to read log:", err)
		return
	}
	fmt.Println(string(data))
}

func addToStash(paths []string) {
	if len(paths) == 0 {
		fmt.Println("Usage: fdrop add <file1> <file2> ...")
		return
	}
	var added []string
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		name := filepath.Base(absPath)
		appendToStash(name, absPath)
		added = append(added, name)
	}
	if len(added) > 0 {
		logAction("Added to stash: " + strings.Join(added, ", "))
	}
}

func appendToStash(name, absPath string) {
	f, _ := os.OpenFile(stashFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s|%s\n", name, absPath))
}

func showStash() {
	items := readStash()
	if len(items) == 0 {
		fmt.Println("Stash is empty.")
		return
	}
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item.Name)
	}
}

func readStash() []StashItem {
	var items []StashItem
	f, err := os.Open(stashFile)
	if err != nil {
		return items
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "|", 2)
		if len(parts) == 2 {
			items = append(items, StashItem{parts[0], parts[1]})
		}
	}
	return items
}

func writeStash(items []StashItem) {
	f, _ := os.Create(stashFile)
	defer f.Close()
	for _, item := range items {
		f.WriteString(fmt.Sprintf("%s|%s\n", item.Name, item.Path))
	}
}

func pasteFiles(args []string, isMove bool) {
	stash := readStash()
	if len(stash) == 0 {
		fmt.Println("Stash is empty.")
		return
	}

	var selected []StashItem
	var targetDir = "."

	// Determine if indices or names are used
	if len(args) == 0 {
		selected = stash
	} else {
		// Last argument could be a target dir
		if len(args) > 1 && !strings.HasPrefix(args[len(args)-1], "stash") && !isStashIndex(args[len(args)-1]) {
			targetDir = args[len(args)-1]
			args = args[:len(args)-1]
		}
		if len(args) >= 1 && args[0] == "stash" {
			selected = selectByIndex(args[1:], stash)
		} else {
			selected = selectByName(args, stash)
		}
	}

	if len(selected) == 0 {
		fmt.Println("No files matched in stash.")
		return
	}

	var success []string
	var remaining []StashItem
	skip := make(map[string]bool)

	for _, item := range selected {
		dest := filepath.Join(targetDir, item.Name)
		var err error
		if isMove {
			err = os.Rename(item.Path, dest)
		} else {
			err = copyPath(item.Path, dest)
		}
		if err != nil {
			fmt.Printf("Failed: %s (%v)\n", item.Name, err)
			continue
		}
		fmt.Printf("Pasted: %s\n", item.Name)
		success = append(success, item.Name)
		skip[item.Path] = true
	}

	for _, item := range stash {
		if !skip[item.Path] {
			remaining = append(remaining, item)
		}
	}
	writeStash(remaining)

	action := "Copied"
	if isMove {
		action = "Moved"
	}
	logAction(fmt.Sprintf("%s: %s", action, strings.Join(success, ", ")))
}

func isStashIndex(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func selectByName(names []string, stash []StashItem) []StashItem {
	var selected []StashItem
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, item := range stash {
		if nameSet[item.Name] {
			selected = append(selected, item)
		}
	}
	return selected
}

func selectByIndex(indices []string, stash []StashItem) []StashItem {
	var selected []StashItem
	for _, idxStr := range indices {
		var idx int
		fmt.Sscanf(idxStr, "%d", &idx)
		if idx >= 1 && idx <= len(stash) {
			selected = append(selected, stash[idx-1])
		}
	}
	return selected
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, srcInfo.Mode())
	}
	return nil
}

