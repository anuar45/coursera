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
	wg := &sync.WaitGroup{}
	m := new(sync.Mutex)
LOOP:
	for {
		select {
		case data := <-in:
			wg.Add(1)
			dataStr := strconv.Itoa(data.(int))
			pd := PipeData{Value: dataStr}
			go CalcSingleHash(&pd, out, wg, m)
		case <-time.After(5 * time.Millisecond):
			break LOOP
		}
	}
	wg.Wait()
	close(out)
}

// MultiHash creates 6 hashes from one input
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		//fmt.Println("DEBUG: Entered chan cycle")
		wg.Add(1)
		pd, ok := data.(PipeData)
		if !ok {
			fmt.Printf("MHASH: Type assertion error. Got: %v\n", data)
		}
		go CalcMultiHash(&pd, out, wg)
	}

	wg.Wait()
	close(out)
}

// CombineResults concatenates result with "_" separator
func CombineResults(in, out chan interface{}) {
	var result string
	var sl []string
	for data := range in {
		pd, ok := data.(PipeData)
		if !ok {
			fmt.Printf("COMBINE: Type assertion error. Got: %v\n", data)
		}
		sl = append(sl, pd.MHashes)
	}
	sort.SliceStable(sl, func(i, j int) bool { return sl[i] < sl[j] })
	result = strings.Join(sl, "_")
	fmt.Println("CombineResults", result)
	out <- result
	close(out)
}

// CalcSingleHash calculates SingleHash value for one data entry
func CalcSingleHash(pd *PipeData, out chan interface{}, wg *sync.WaitGroup, m *sync.Mutex) {
	waitCh := make(chan struct{})
	fmt.Println(pd.Value, "SingleHash", "data", pd.Value)
	m.Lock()
	md5Hash := DataSignerMd5(pd.Value)
	m.Unlock()
	fmt.Println(pd.Value, "SingleHash", "md5(data)", md5Hash)
	pd.Lock()
	pd.CrcMd5Hash = DataSignerCrc32(md5Hash)
	pd.Unlock()
	fmt.Println(pd.Value, "SingleHash", "crc32(md5(data))", pd.CrcMd5Hash)

	go func() {
		pd.Lock()
		pd.CrcHash = DataSignerCrc32(pd.Value)
		pd.Unlock()
		fmt.Println(pd.Value, "SingleHash", "crc32(data)", pd.CrcHash)
		waitCh <- struct{}{}
	}()

	<-waitCh

	pd.TildaHash = pd.CrcHash + "~" + pd.CrcMd5Hash
	fmt.Println(pd.Value, "SingleHash", "result", pd.TildaHash)
	out <- *pd
	wg.Done()
}

// CalcMultiHash calculates hashes in interation
func CalcMultiHash(pd *PipeData, out chan interface{}, wg *sync.WaitGroup) {
	wg2 := &sync.WaitGroup{}
	thNum := 6
	pd.MHash = make(map[int]string)
	for th := 0; th < thNum; th++ {
		wg2.Add(1)
		go func(th int) {
			thStr := strconv.Itoa(th)
			crcStr := DataSignerCrc32(thStr + pd.TildaHash)
			pd.Lock()
			pd.MHash[th] = crcStr
			pd.Unlock()
			fmt.Println(pd.TildaHash, "MultiHash: crc32(th+step1))", thStr, crcStr)
			wg2.Done()
		}(th)
	}

	wg2.Wait()

	for i := 0; i < thNum; i++ {
		pd.MHashes += pd.MHash[i]
	}

	fmt.Println(pd.TildaHash, "MultiHash result:", pd.MHashes)
	out <- *pd
	wg.Done()
}
