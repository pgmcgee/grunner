package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	flag.Parse()

	cmd := flag.Arg(0)

	if cmd != "run" {
		fmt.Println("You must specify to run a file")
		os.Exit(0)
	}

	runFile := flag.Arg(1)

	if runFile == "" {
		fmt.Println("You must specify a file to run")
		os.Exit(0)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			watcher.Add(path)
			fmt.Println("Watching", path)
		}

		return nil
	})

	done := make(chan bool)
	go func() {
		for {
			goRunCommand := exec.Command("go", "run", runFile)

			stdout, err := goRunCommand.StdoutPipe()
			if err != nil {
				log.Println("Error grabbing stdout", err)
			}
			stderr, err := goRunCommand.StderrPipe()
			if err != nil {
				log.Println("Error grabbing stderr", err)
			}

			goRunCommand.Start()

		ProcessEvents:
			for {
				// Colorize Green
				fmt.Print("\033[32m")
				io.Copy(os.Stdout, stdout)
				fmt.Print("\033[0m")

				// Colorize Red
				fmt.Print("\033[31m")
				io.Copy(os.Stderr, stderr)
				fmt.Print("\033[0m")

				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("Modified File, restarting: ", event.Name)
						break ProcessEvents
					}
				case err := <-watcher.Errors:
					log.Println("Error: ", err)
				}
			}

			goRunCommand.Process.Kill()
		}
	}()

	<-done
}
