package main
import(
"fmt"
"time")

//select is used to wait on channel operations.
//this allows you to listen for your channels' finishing their tasks when they're done
//the following code listens to c1 and c2, printing right when they get filled, causing it to be on time.

func main() {
	c1 := make(chan string)
	c2 := make(chan string)
	go func() {
		time.Sleep(1 * time.Second)
		c1 <- "one"
	}()
	go func() {
		time.Sleep(2 * time.Second)
		c2 <- "two"
	}()
}