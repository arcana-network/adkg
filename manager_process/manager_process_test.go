package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// test if the main process will create a new child process
// on receiving MSG_TO_MANAGER:DPSS_START
func TestCreatingNewNode(t *testing.T) {
	defer os.Remove("./testStart")
	// compile the mock binary
	cmd := exec.Command("go", "build",
		"-o", "./testStart",
		"./test-util/mock-dpss-start-binary",
	)
	cmd.Start()
	cmd.Wait()

	assert.Equal(t, len(runningNodes), 0)

	// run a test binary that print MSG_TO_MANAGER:DPSS_START
	go startNewNode("./testStart", "./new-node-config")
	time.Sleep(5 * time.Second)
	//must create more than one child_processes
	assert.Greater(t, len(runningNodes), 1)
}

// test if the main process will output the correct msg
// on receiving MSG_TO_MANAGER:DPSS_END
func TestEndingDPSS(t *testing.T) {
	scanner, reader, writer := mockLogger(t)
	defer resetLogger(reader, writer)
	defer os.Remove("./testEnd")
	// compile the mock binary
	cmd := exec.Command("go", "build",
		"-o", "./testEnd",
		"./test-util/mock-dpss-end-binary",
	)
	cmd.Start()
	cmd.Wait()

	var msg string
	msgCh := make(chan string, 1)
	go func(msgCh chan string) {
		for {
			scanner.Scan()
			msg = scanner.Text()
			msgCh <- msg
		}
	}(msgCh)

	// run a test binary that print MSG_TO_MANAGER:DPSS_END
	go startNewNode("./testEnd", "./new-node-config")

	timeout := time.NewTimer(4 * time.Second)
	for {
		select {
		case <-timeout.C:
			assert.Contains(t, msg, "Received DPSS ending signal from child")
			return
		case msg = <-msgCh:
			continue
		}
	}
}

// test if the main process will kill the child process
// on receiving MSG_TO_MANAGER:KILL_NODE
func TestKillingProcess(t *testing.T) {
	defer os.Remove("./testKill")
	// compile the mock binary
	cmd := exec.Command("go", "build",
		"-o", "./testKill",
		"./test-util/mock-kill-process-binary",
	)
	cmd.Start()
	cmd.Wait()

	// assert zero running nodes before starts
	assert.Equal(t, len(runningNodes), 0)
	// run a test binary that print MSG_TO_MANAGER:KILL_NODE
	go startNewNode("./testKill", "./new-node-config")
	time.Sleep(1 * time.Second)

	// assert created one running node
	assert.Equal(t, len(runningNodes), 1)
	time.Sleep(7 * time.Second)
	// assert the running node has been killed
	assert.Equal(t, len(runningNodes), 0)

}

// set log output to a mock logger so we can check in test
func mockLogger(t *testing.T) (*bufio.Scanner, *os.File, *os.File) {
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	log.SetOutput(writer)

	return bufio.NewScanner(reader), reader, writer
}

// reset the log output to stderr
func resetLogger(reader *os.File, writer *os.File) {
	err := reader.Close()
	if err != nil {
		fmt.Println("error closing reader was ", err)
	}
	if err = writer.Close(); err != nil {
		fmt.Println("error closing writer was ", err)
	}
	log.SetOutput(os.Stderr)
}
