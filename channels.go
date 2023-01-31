package main

import "fmt"

/*
Channels
--> pipes to send data to another thread / routine.
--> sending end automatically blocks until receiving end is ready
--> helps goroutines synchronize data without explicit locks
--> buffered channels
		--> maximum size of a channel
		--> only sends when the buffer is full
		--> receive requests block when the buffer is empty.
		--> overfill: throws an error!
*/

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c) //owns the channel first
	go sum(s[len(s)/2:], c) //gets the channel once the first routine finishes
	x, y := <-c, <-c        // receive from c
	fmt.Println(x, y, x+y)

	//using a buffer
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	//ch <- 3 overfilling the buffer causes an error
	fmt.Println(<-ch)
	fmt.Println(<-ch)

	//directional channels increase type-safety of programs.
	//channel behavior is declared in a function's arguments

	pings := make(chan string, 1)
	pongs := make(chan string, 1)

	ping(pings, "hey there") //goes into pings
	pong(pings, pongs)       //pings goes local variable goes into pongs
	fmt.Println(<-pongs)
}
func ping(pings chan<- string, msg string) { //only receives
	pings <- msg //message goes into pings
}
func pong(pings <-chan string, pongs chan<- string) {
	msg := <-pings //pings goes into this function's local msg
	pongs <- msg   //msg goes into pongs
}
