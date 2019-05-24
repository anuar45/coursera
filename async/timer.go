package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(10 * time.Second)

	fmt.Println(<-timer.C)
}
