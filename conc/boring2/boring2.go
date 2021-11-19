/*
	Fan-in Mod
	扇入模式，多个chnnel连接到一个channel；可以类比消息队列

	注意，这里如果用列表代替两个参数，则可能会看不到应有的多个routine争抢打印的效果，
	而是发现main倾向于去等待某个routine的消息，这个原因还未知。

	用多个参数而非列表参数，则不会出现这个问题。
*/
package main

import (
    "fmt"
    "time"
    "math/rand"
)

func boring(msg string) <-chan string { // Returns receive-only channel of strings.
    c := make(chan string)
    go func() { // We launch the goroutine from inside the function.
        for i := 0; ; i++ {
            c <- fmt.Sprintf("%s %d", msg, i)
            time.Sleep(time.Duration(rand.Intn(2e3)) * time.Millisecond)
        }
    }()
    return c // Return the channel to the caller.
}

// func fanIn1(inputs *[]<-chan string) <-chan string {
// 	c := make( chan string)
// 	for _, input := range *inputs {
// 		fmt.Printf("Fanned: %v\n", input)
// 		go func() { for { c <- <-input } }()
// 	}
// 	return c
// }

func fanIn2(input1, input2 <-chan string) <-chan string{
	c := make(chan string)
    go func() { for { c <- <-input1 } }()
    go func() { for { c <- <-input2 } }()
    return c
}

func main() {
	// channels := []<-chan string{
	// 	boring("i"), boring("j"),
	// 	boring("i"), boring("j"),
	// 	boring("i"), boring("j"),
	// 	boring("i"), boring("j"),
	// 	boring("i"), boring("j"),
	// }
	// c := fanIn1( &channels)

	c := fanIn2(boring("A"), boring("B"))

	for i := 0; i < 10; i++ {
        fmt.Println(<-c)
    }

    fmt.Println("Leaving.")
}