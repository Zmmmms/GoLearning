package main

/*
	3-routine交叉打印
*/

import (
	"fmt"
	// "time"
	// "math/rand"
)

func output(start int, c chan int, condition, signal chan bool){
	go func() {
		for ; ; start+=3{
			<-condition
			c <- start
			signal <- true
		}
	}()
}

func main() {
	c := make( chan int)
	signals1 := make( chan bool)
	signals2 := make( chan bool)
	signals3 := make( chan bool)

	go output(0, c, signals1, signals2)
	go output(1, c, signals2, signals3)
	go output(2, c, signals3, signals1)

	signals1 <- true
	for i := 0; i<100; i++{
		fmt.Println( <-c)
	}
}
