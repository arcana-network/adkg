package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"sync"
)

// ANSI color codes
const (
	Magenta = "\033[35m"
	Purple  = "\033[34m"
	Reset   = "\033[0m" // Reset to default terminal color
)

// printColoredOutput reads from the given reader (stdout or stderr) and prints each line with the specified color
func printColoredOutput(r io.Reader, color string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(color + scanner.Text() + Reset)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading output: %s", err)
	}
}

// startInstance starts a child process and streams its output in a specific color
func startInstance(instance int, wg *sync.WaitGroup, color string) {
	defer wg.Done()

	// "start" is the command that should start a node in adkg and all the necessary services
	// when sending another command such as "version" the process will print and exit
	// uncomment the following line to try that out (it will print the version in 2 different colors and shut down the main process)
	//cmd := exec.Command("../arcana_impl_adkg/myapp", "version")
	config_file := "./local-setup-data/config.test." + strconv.Itoa(instance+1) + ".json"
	secret_config_file := "./local-setup-data/config.test." + strconv.Itoa(instance+1) + ".json"
	cmd := exec.Command("./myapp", "start",
		"--config", config_file,
		"--secret-config", secret_config_file,
	)

	// Stream the output directly instead of capturing it
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Instance %d: Error obtaining stdout: %s", instance, err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Instance %d: Error obtaining stderr: %s", instance, err)
		return
	}

	// Sned message to child process
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Instance %d: Error obtaining stdin: %s", instance, err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Instance %d: Error starting command: %s", instance, err)
		return
	}

	go printColoredOutput(stdoutPipe, color)
	go printColoredOutput(stderrPipe, color)

	// Send to child
	stdinPipe.Write([]byte("Hello from main!\n"))

	if err := cmd.Wait(); err != nil {
		log.Printf("Instance %d: Command finished with error: %s", instance, err)
	}
}

func main() {
	var wg sync.WaitGroup

	colors := []string{Magenta, Purple}

	// Start two instances of akdg-dpss process
	// If all is well they both start up and keep running
	// But right now, only 1 process can survive at a time (probably the port collision)
	// and the other one keeps showing errors.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go startInstance(i, &wg, colors[i%len(colors)])
	}

	wg.Wait()

	log.Println("Shutting down...")
}
