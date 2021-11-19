package main

/*
	基于信号量的routine控制；
*/

import (
	"fmt"
	"time"
	"math/rand"
)

type Message struct {
    str string
    wait chan bool
}

func boring(msg string) <-chan string { // Returns receive-only channel of strings.
    c := make(chan string)
    go func() { // We launch the goroutine from inside the function.
        for i := 0; ; i++ {
            c <- fmt.Sprintf("%s %d", msg, i)
            time.Sleep(time.Duration(rand.Intn(4e3)) * time.Millisecond)
        }
    }()
    return c // Return the channel to the caller.
}

func fanIn(input1, input2 <-chan string) <-chan Message {
	waitForIt := make( chan bool)
	c := make( chan Message)
	go func() { 
		for { 
			c <- Message{<-input1, waitForIt} 
			<-waitForIt 
		} 
	} ()
	go func() { for { c <- Message{<-input2, waitForIt}; <-waitForIt } } ()
	return c
}

func main() {
	c := fanIn( boring("A"), boring("B"))
	for i := 0; i<5; i++ {
		msg1 := <-c; fmt.Println(msg1.str)
		msg2 := <-c; fmt.Println(msg2.str)
		msg1.wait <- true
		msg2.wait <- true
	}
}