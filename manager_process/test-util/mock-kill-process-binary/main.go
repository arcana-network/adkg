package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(5 * time.Second)
	fmt.Println("MSG_TO_MANAGER:KILL_NODE")
	time.Sleep(3 * time.Second)
}
