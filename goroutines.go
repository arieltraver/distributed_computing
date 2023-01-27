package main

/*
go routine: a lightweight thread.
--> lightweight: uses the SAME ADDRESS SPACE as current process
go(f,x,y,z)
--> f,x,y,z evaluated in the current routine
--> f(x,y,z) executed in the new routine
*/
import (
	"fmt"
	"time"
)

func say(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func main() {
	go say("world")
	say("hello")
}
