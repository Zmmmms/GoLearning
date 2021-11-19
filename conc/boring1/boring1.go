package main

/*
    Generator：function that returns a ch
    生成式的写法

    也可以换个角度理解，这样的写法相当于我们通过c调用了boring服务，
*/

import (
    "fmt"
    "time"
)

func main(){
    c := boring("boring!") // Function returning a channel.
    for i := 0; i < 5; i++ {
        fmt.Printf("You say: %q\n", <-c)
    }
    fmt.Println("You're boring; I'm leaving.")
}

func boring(msg string) <-chan string { // Returns receive-only channel of strings.
    c := make(chan string)
    go func() { // We launch the goroutine from inside the function.
        for i := 0; ; i++ {
            c <- fmt.Sprintf("%s %d", msg, i)
            time.Sleep(time.Duration(1) * time.Second)
        }
    }()
    return c // Return the channel to the caller.
}