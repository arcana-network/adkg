package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// ANSI color codes
const (
	Magenta = "\033[35m"
	Purple  = "\033[34m"
	Reset   = "\033[0m" // Reset to default terminal color
)

const (
	MSG_PREFIX     = "MSG_TO_MANAGER:"
	MSG_DPSS_START = "DPSS_START"
	MSG_DPSS_END   = "DPSS_END"
	MSG_SHUT_DOWN  = "KILL_NODE"
	BIN_NAME       = "./myapp"
	CONFIG_PATH    = "./manager_process/new-node-config"
)

var colorFlip bool
var runningNodes = make(map[int]*exec.Cmd)

// createNodeInstance creates and return a start command
// of the child process
func createNodeCmd(binName string, configPath string) (*exec.Cmd, error) {

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		log.Error("manager process err: config file not exist or more than one")
		return nil, err
	}
	config_file := fmt.Sprintf("%s/%s", configPath, files[0].Name())
	// TODO: what should be in the secret_config_file
	secret_config_file := config_file

	cmd := exec.Command(binName, "start",
		"--config", config_file,
		"--secret-config", secret_config_file,
	)

	return cmd, nil
}

// processOutput scans the io.Reader of a process and
// sends the msg to nodeCh if it contains the MSG_PREFIX
// It also prints the msg with the specified color
func processOutput(r io.Reader, color string, nodeCh chan string, pid int) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		_, msg, found := strings.Cut(scanner.Text(), MSG_PREFIX)
		if found {
			nodeCh <- strings.TrimSpace(msg)
		}
		fmt.Println(color + strconv.Itoa(pid) + " " + scanner.Text() + Reset)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading output: %s\n", err)
	}
}

// on receiving ending signal, send SIGINT to all the
// child processes and wait for them to stop
func stopOnInterrupt() {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	osSig := <-osSignal
	log.Println("Termination started, signal: " + osSig.String())
	for _, v := range runningNodes {
		v.Process.Signal(syscall.SIGINT)
		v.Process.Wait()
	}
}

// nextColor returns alternating colors
func nextColor() (color string) {
	if colorFlip {
		colorFlip = !colorFlip
		return Magenta
	}
	colorFlip = !colorFlip
	return Purple
}

// startNewNode creates a new child process and monitors
// its output to create or kill a process
func startNewNode(binName string, cfgPath string) {
	nodeCh := make(chan string)

	cmd, err := createNodeCmd(binName, cfgPath)
	if err != nil {
		log.Errorf("Error creating child process: %s\n", err)
		return
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error obtaining stdout: %s\n", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error obtaining stderr: %s\n", err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s\n", err)
		return
	}
	log.Printf("Starting Node with PID: %v\n", cmd.Process.Pid)
	runningNodes[cmd.Process.Pid] = cmd

	color := nextColor()
	go processOutput(stdoutPipe, color, nodeCh, cmd.Process.Pid)
	go processOutput(stderrPipe, color, nodeCh, cmd.Process.Pid)

	for {
		msg := <-nodeCh
		switch msg {
		case MSG_DPSS_START:
			// DPSS starting, create a new node
			log.Println("Creating a new node")
			go startNewNode(binName, cfgPath)
		case MSG_DPSS_END:
			// DPSS ends
			log.Println("Received DPSS ending signal from child")
		case MSG_SHUT_DOWN:
			// Epoch change
			log.Println("Epoch changed, terminating old node")
			delete(runningNodes, cmd.Process.Pid)
			cmd.Process.Kill()
		}

	}
}

// TODO: add functions to manually start/stop a process
func main() {
	go startNewNode(BIN_NAME, CONFIG_PATH)
	stopOnInterrupt()
	log.Println("Shutting down...")

}
