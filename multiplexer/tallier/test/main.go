package main

import (
	"github.com/Jille/bindlink/multiplexer/tallier"

	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	t := tallier.New(500, 5000)
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			fmt.Printf("%d %v\n", t.Count(), t)
			time.Sleep(500 * time.Millisecond)
		}
	}()
	for {
		t.Tally()
		reader.ReadString('\n')
	}
}
