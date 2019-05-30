package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PipeData struct {
	sync.Mutex
	Value      string
	CrcHash    string
	CrcMd5Hash string
	TildaHash  string
	MultiHash  []string
}

// ExecutePipeline makes pipiline of funcs
func ExecutePipeline(jobFuncs ...job) {
	in := make(chan interface{})
	out := make(chan interface{})
	for i := 0; i < len(jobFuncs)-1; i++ {
		out := make(chan interface{})
		go jobFuncs[i](in, out)
		in = out
	}
	jobFuncs[len(jobFuncs)-1](in, out)
}

// SingleHash accepts in channel in makes operations and send to out
func SingleHash(in, out chan interface{}) {
	var result string
	pch := make(chan PipeData)
LOOP:
	for {
		select {
		case data := <-in:
			dataStr := strconv.Itoa(data.(int))
			pd := PipeData{Value: dataStr}
			go CalcSingleHash(pd, pch)
		case <-time.After(5 * time.Millisecond):
			out <- "end"
			break LOOP
		}
	}
}

// MultiHash creates 6 hashes from one input
func MultiHash(in, out chan interface{}) {
	var result string
	for data := range in {
		result = ""
		dataStr := data.(string)
		if dataStr != "end" {
			for th := 0; th < 6; th++ {
				thStr := strconv.Itoa(th)
				crcStr := DataSignerCrc32(thStr + dataStr) // Paralell
				result += crcStr
				fmt.Println(dataStr, "MultiHash: crc32(th+step1))", th, crcStr)
			}
			fmt.Println(dataStr, "MultiHash result:", result)
			out <- result
		} else {
			out <- dataStr
			break
		}
	}
}

// CombineResults concatenates result with "_" separator
func CombineResults(in, out chan interface{}) {
	var result string
	var sl []string
	for data := range in {
		dataStr := data.(string)
		if dataStr != "end" {
			sl = append(sl, dataStr)
		} else {
			sort.SliceStable(sl, func(i, j int) bool { return sl[i] < sl[j] })
			result = strings.Join(sl, "_")
			fmt.Println("CombineResults", result)
			break
		}
	}
	out <- result
}

// CalcSingleHash calculates SingleHash value for one data entry
func CalcSingleHash(pd PipeData, pch chan PipeData) {
	fmt.Println(pd.Value, "SingleHash", "data", pd.Value)
	md5Hash := DataSignerMd5(pd.Value) // Lock
	fmt.Println(pd.Value, "SingleHash", "md5(data)", md5Hash)
	pd.CrcMd5Hash = DataSignerCrc32(md5Hash) // Paralell
	fmt.Println(pd.Value, "SingleHash", "crc32(md5(data))", pd.CrcMd5Hash)

	// sync this with parent
	go func() {
		pd.CrcHash = DataSignerCrc32(pd.Value) // Paralell
		fmt.Println(pd.Value, "SingleHash", "crc32(data)", pd.CrcHash)
	}()

	pd.TildaHash = pd.CrcHash + "~" + pd.CrcMd5Hash
	fmt.Println(pd.Value, "SingleHash", "result", pd.TildaHash)
}
