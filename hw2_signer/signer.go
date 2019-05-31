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
	MHash      map[int]string
	MHashes    string
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
	wg := new(sync.WaitGroup)
LOOP:
	for {
		select {
		case data := <-in:
			wg.Add(1)
			dataStr := strconv.Itoa(data.(int))
			pd := PipeData{Value: dataStr}
			go CalcSingleHash(&pd, out, wg)
		case <-time.After(5 * time.Millisecond):
			break LOOP
		}
	}
	wg.Wait()
	out <- struct{}{}
}

// MultiHash creates 6 hashes from one input
func MultiHash(in, out chan interface{}) {
	wg := new(sync.WaitGroup)
	for data := range in {
		wg.Add(1)
		pd, ok := data.(PipeData)
		if ok {
			go CalcMultiHash(&pd, out, wg)
		} else {
			break
		}
	}
	wg.Wait()
	out <- struct{}{}
}

// CombineResults concatenates result with "_" separator
func CombineResults(in, out chan interface{}) {
	var result string
	var sl []string
	for data := range in {
		pd, ok := data.(PipeData)
		if ok {
			sl = append(sl, pd.MHashes)
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
func CalcSingleHash(pd *PipeData, out chan interface{}, wg *sync.WaitGroup) {
	waitCh := make(chan struct{})
	fmt.Println(pd.Value, "SingleHash", "data", pd.Value)
	md5Hash := DataSignerMd5(pd.Value) // Lock
	fmt.Println(pd.Value, "SingleHash", "md5(data)", md5Hash)
	pd.CrcMd5Hash = DataSignerCrc32(md5Hash)
	fmt.Println(pd.Value, "SingleHash", "crc32(md5(data))", pd.CrcMd5Hash)

	go func() {
		pd.CrcHash = DataSignerCrc32(pd.Value)
		fmt.Println(pd.Value, "SingleHash", "crc32(data)", pd.CrcHash)
		waitCh <- struct{}{}
	}()

	<-waitCh

	pd.TildaHash = pd.CrcHash + "~" + pd.CrcMd5Hash
	fmt.Println(pd.Value, "SingleHash", "result", pd.TildaHash)
	out <- pd
	wg.Done()
}

// CalcMultiHash calculates hashes in interation
func CalcMultiHash(pd *PipeData, out chan interface{}, wg *sync.WaitGroup) {
	wg2 := new(sync.WaitGroup)
	thNum := 6
	for th := 0; th < thNum; th++ {
		wg2.Add(1)
		thStr := strconv.Itoa(th)
		go func() {
			crcStr := DataSignerCrc32(thStr + pd.TildaHash)
			pd.MHash[th] = crcStr
			fmt.Println(pd.TildaHash, "MultiHash: crc32(th+step1))", th, crcStr)
			wg2.Done()
		}()
	}

	wg2.Wait()

	for i := 0; i < thNum; i++ {
		pd.MHashes += pd.MHash[i]
	}

	fmt.Println(pd.TildaHash, "MultiHash result:", pd.MHashes)
	out <- pd
	wg.Done()
}
