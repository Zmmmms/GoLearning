# Go Conc Pattern

根据GoConcrrencyPattern总结多个场景下的Go并发模式；

## 基础：goroutine / go关键字
```go
go boring("hello")
```
启动一个`goroutine`来执行函数；
值得注意的是，和其他语言的多线程模型类似，如果主线程不等待它则会立刻结束了；

## 基础：channel / chan关键字
利用`chan`来通信；
chan的设计很特别，在我理解中是一个只写结构的队列。这意味着chan中的消息不拿出来是看不到消息体的。

我们可以基于此先做一个基本的生产-消费者模型；
boring函数中不断地生产消息，main端则消费五次。由于生产的时延大于消费时延，消费者自然会因为没有消息可消费而阻塞，即读空chan是会阻塞的。

```go
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
```

## 基础：Buffered Channel

以Buffer的方式实现非阻塞；但该实践目前掠过此部分。

## 模式：生成式

返回chan的函数！注意上一段代码，我们用chan作为参数，然后在消费方注册实例去接入它。
相比这种方式，我们可能更想要的是一种服务调用的感觉：消费者在需要时调用生产者。

我们利用闭包来将boring做成一个服务，其返回的chan作为该服务的句柄；消费者在需要时调用即可：

```go
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

func main(){
    c := boring("boring!") // Function returning a channel.
    for i := 0; i < 5; i++ {
        fmt.Printf("You say: %q\n", <-c)
    }
    fmt.Println("You're boring; I'm leaving.")
}
```

## 模式：层叠

考虑一个很常用的场景：并行调用；
我们可能会同时调用多个服务来请求同一个query；如搜索时，我们可能会同时搜索baidu、google、bing，使用哪个结果只取决于哪个服务先响应，这样就保证了我们结果的最低时延。

对应这个模式，我们并行调用两个boring服务，消费方无差别消费两者任意的消息；为了使其也对应生成式模型，我们再编写一个服务，该服务`fan-in`两个boring并提供一个chan作为句柄以供消费方调用。

```go
func boring(msg string) <-chan string { same as previous }

func fanIn(input1, input2 <-chan string) <-chan string {
    c := make(chan string)
    go func() { for { c <- <-input1 } }()
    go func() { for { c <- <-input2 } }()
    return c
}

func main() {
    c := fanIn(boring("Joe"), boring("Ann"))
    for i := 0; i < 10; i++ {
        fmt.Println(<-c)
    }
    fmt.Println("You're both boring; I'm leaving.")
}
```

贴心的图示

<img src="/Users/zms/Desktop/截屏2021-11-20 下午3.39.58.png" alt="截屏2021-11-20 下午3.39.58" style="zoom: 33%;" />



## 模式：调用间同步

调用间同步也是很常见的场景；我们的工作可能分为两步，第二步工作需要等待第一步的多个并行工作全部Join才可以开始。

考虑两个`boring`服务，其均输出序列消息b1、b2、b3；但我们想使所有boring服务生产的b1被消费后才开始生产b2。

我们使用一个消息信号量`wait chan`来完成；每条消息都将接入一个wait chan，只有其被消费后，生产者才开始生产下一条消息。这样，我们在消费者处同时给出wait信号，表示该阶段的消息已被消费完成，让生产者开始下一阶段的生产。

```go
type Message struct {
    str string
    wait chan bool
}

func boring(msg string) <-chan Message {
    c := make(chan Message)
    go func() { 
        for i := 0; ; i++ {
          	waitchan := make(chan bool)
	          c <- Message{fmt.Sprintf("%s %d", msg, i), waitchan}
            time.Sleep(time.Duration(rand.Intn(4e3)) * time.Millisecond)
        }
    }()
    return c // Return the channel to the caller.
}

func fanIn(input1, input2 <-chan Message) <-chan Message {
  same as previous
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
```

## 基础：Select

Fan-In多个服务的代码会显得易读性稍差：

```go
    go func() { for { c <- <-input1 } }()
    go func() { for { c <- <-input2 } }()
```

我们更倾向于体现出一种“哪个chan可读就读哪个”的感觉而非开启两个routine去读。

以`select`的方式来完成：

```go
go func() {
        for {
            select {
            case s := <-input1:  c <- s
            case s := <-input2:  c <- s
            }
        }
    }()
```

select只能在并行情况中使用，由于并行程序的特性，select中的case是随机而非顺序执行的。

## 模式：超时

超时的场景有很多。如调用某服务时希望给到调用方一个最长超时时限，超过即返回。

Go中的超时一般使用时间类的服务ch完成。如下面代码中，消费者会在5秒后结束消费。

```go
c := boring("Joe")
timeout := time.After(5 * time.Second)
for {
  select {
    case s := <-c:
    	fmt.Println(s)
    case <-timeout:
    	fmt.Println("You talk too much.")
    return
  }
}
```

或控制生产者停止生产：

```go
select {
  case c <- fmt.Sprintf("%s: %d", msg, i):
  	// pass
  case <-quit:
  	return
}
```

## 模式：链式

链式传递消息，层层阻塞。不太算单独的一个模式；

以下创建一个这样的结构：每层会将上层的消息值加1后传递给下层

```go
func f(left, right chan int) {
    left <- 1 + <-right
}

func main() {
    const n = 10000
    leftmost := make(chan int)
    right := leftmost
    left := leftmost
    for i := 0; i < n; i++ {
        right = make(chan int)
        go f(left, right)
        left = right
    }
    go func(c chan int) { c <- 1 }(right)
    fmt.Println(<-leftmost)
}
```

## 例：Google搜索

本节给出了执行对一个keyword进行搜索时的模式；
搜索时，引擎应当在多个模块对关键字进行搜索，如网站、图片、视频模块等，这样用户在不同结果间切换就不用等待。

### 串行式

多模块的搜索若为串行，则有如下伪码：

```go
type Search func(query string) Result
type Result string

// 搜索模块范式，以模块名创建搜索模块
func fakeSearch(kind string) Search {
        return func(query string) Result {
              time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
              return Result(fmt.Sprintf("%s result for %q\n", kind, query))
        }
}

// 假定有如下三个搜索模块，每次搜索都需要keyword在三个模块中的结果
var (
    Web = fakeSearch("web")
    Image = fakeSearch("image")
    Video = fakeSearch("video")
)

func Google(query string) (results []Result) {
    // 串行搜索不同模块中的结果
    results = append(results, Web(query))
    results = append(results, Image(query))
    results = append(results, Video(query))
    return
}
```

### 并行式

模块若为并行，则我们可以稍微改写即成：

```go
func Google(query string) (results []Result) {
  c := make( chan Result)
	go func() { c <- Web(query) }()
	go func() { c <- Image(query) }()
	go func() { c <- Video(query) }()

	for i := 0; i < 3; i++ {
		results = append( results, <-c)
	}
  return
}
```

### 加入总线超时

响应过慢时，需要给整个搜索过程加入超时

```go
c := make(chan Result)
go func() { c <- Web(query) } ()
go func() { c <- Image(query) } ()
go func() { c <- Video(query) } ()

timeout := time.After(80 * time.Millisecond)
for i := 0; i < 3; i++ {
  select {
    case result := <-c:
    results = append(results, result)
    case <-timeout:
    fmt.Println("timed out")
    return
  }
}
return
```

### 分布式搜索

事实上，我们对每个模块的搜索也不只是一个单独的过程，而是将一份query复制多份后请求多个相同服务来保证容错；对结果，我们取最先拿到的响应。这就是之前的fan-in模式的应用。

我们编写fan-in来扇入对多个服务器执行同一个query的搜索结果。当然，仍带有总线超时。

```go
func Fanin(query string, replicas ...Search) Result {
    c := make(chan Result)
    searchReplica := func(i int) { c <- replicas[i](query) }
    for i := range replicas {
        go searchReplica(i)
    }
    return <-c
}

func Google(query string) (results []Result) {
  c := make(chan Result)
  go func() { c <- First(query, Web1, Web2) } ()
  go func() { c <- First(query, Image1, Image2) } ()
  go func() { c <- First(query, Video1, Video2) } ()
  timeout := time.After(80 * time.Millisecond)
  for i := 0; i < 3; i++ {
    select {
      case result := <-c:
      results = append(results, result)
      case <-timeout:
      fmt.Println("timed out")
      return
    }
  }
  return
}
```

均达成：“没有回调、没有condition、没有锁“

## 结语

go在原生层直接用原语支持线程内并发，非常精妙灵活。
