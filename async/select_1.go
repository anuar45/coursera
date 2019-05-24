package main

import (
	"fmt"
)

func main() {
	ch1 := make(chan int, 1)
	ch2 := make(chan int)
	// ch1 <- 1
	// go func(in chan int) {
	// 	fmt.Println(<-in)
	// }(ch2)

	select {
	case val := <-ch1:
		fmt.Println("ch1 val", val)
	case <-ch2:
		fmt.Println("put val to ch2")
	default:
		fmt.Println("default case")
	}

}
