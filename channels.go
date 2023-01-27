package main

import "fmt"

/*
Channels
--> pipes to send data to another thread / routine.
--> sending end automatically blocks until receiving end is ready
--> helps goroutines synchronize data without explicit locks
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
}
