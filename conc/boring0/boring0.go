package main

/*
	一个类似于生产-消费者的模型。
	main中的for循环会阻塞等待c通道中有消息可读。
*/
import (
	"fmt"
	"time"
	// "math/rand"
)

func boring(msg string, c chan string) {
	for i := 0; ; i++ {
		c <- fmt.Sprintf("%s %v", msg, i)
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func main() {
	c := make(chan string, 5)
	go boring("zms", c)
	for i := 0; i < 5; i++ {
		fmt.Printf("Blocking here\n")
        fmt.Printf("You say: %q\n", <-c) // Receive expression is just a value.
    }
    fmt.Println("Leave.")
}