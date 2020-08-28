package main

import (
	"fmt"
	"os/exec"

	"bufio"
	"io"
	"strconv"
	"strings"

	"time"
)

// Need to run with sudo because of evtest
// go build test.go && ./test
func main() {
	go readTouchPad("8")

	time.Sleep(1000 * time.Second)
}

func readTouchPad(inputNum string) {
	cmd := exec.Command("evtest")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	r := bufio.NewReader(stdout)
	//go func() {
	io.WriteString(stdin, inputNum)
	stdin.Close()
	//}()

	x, y := -1, -1
	for {
		in, _ := r.ReadString('\n')

		if strings.Contains(in, "(ABS_X),") {
			split := strings.Split(in, " ")
			x, _ = strconv.Atoi(strings.Trim(split[len(split)-1], "\n "))
		} else if strings.Contains(in, "(ABS_Y),") {
			split := strings.Split(in, " ")
			y, _ = strconv.Atoi(strings.Trim(split[len(split)-1], "\n "))
		} else {
			continue
		}

		if x == -1 || y == -1 {
			continue
		}

		doSomething(x, y)
		x, y = -1, -1
	}
}

func doSomething(x, y int) {
	fmt.Println(x, y)
}
