package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ExecutePipeline makes pipiline of funcs
func ExecutePipeline(jobfuncs ...job) {
	in := make(chan interface{})
	for _, jobFunc := range jobfuncs {
		out := make(chan interface{})
		go jobFunc(in, out)
		in = out
	}
	time.Sleep(100 * time.Second)
}

// SingleHash accepts in channel in makes operations and send to out
func SingleHash(in, out chan interface{}) {
	var result string
LOOP:
	for {
		select {
		case data <- in:
			dataStr := strconv.Itoa(data.(int))
			fmt.Println(dataStr, "SingleHash", "data", dataStr)
			md5Hash := DataSignerMd5(dataStr)
			fmt.Println(dataStr, "SingleHash", "md5(data)", md5Hash)
			crcMd5Hash := DataSignerCrc32(md5Hash)
			fmt.Println(dataStr, "SingleHash", "crc32(md5(data))", crcMd5Hash)
			crcHash := DataSignerCrc32(dataStr)
			fmt.Println(dataStr, "SingleHash", "crc32(data)", crcHash)
			result = crcHash + "~" + crcMd5Hash
			fmt.Println(dataStr, "SingleHash", "result", result)
			out <- result
		case <- time.After(5*time.Millisecond):
			break LOOP
			out <- "end"
		}
	}
}

// MultiHash creates 6 hashes from one input
func MultiHash(in, out chan interface{}) {
	var result string
	for data := range in {
		result = ""
		dataStr := data.(string)
		for th := 0; th < 6; th++ {
			thStr := strconv.Itoa(th)
			crcStr := DataSignerCrc32(thStr + dataStr)
			result += crcStr
			fmt.Println(dataStr, "MultiHash: crc32(th+step1))", th, crcStr)
		}
		fmt.Println(dataStr, "MultiHash result:", result)
		out <- result
	}
}

// CombineResults concatenates result with "_" separator
func CombineResults(in, out chan interface{}) {
	var result string
	var sl []string
	for {
		select {
		case data := <-in:
			dataStr := data.(string)
			sl = append(sl, dataStr)
		case 
		}
	}

	result = strings.Join(sl, "_")
	fmt.Println("CombineResults", result)
	out <- result
}
