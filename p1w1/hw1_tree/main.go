package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(w io.Writer, p string, pf bool) error {
	var prefixStr string
	dirList, _ := ioutil.ReadDir(p)

	for _, dir := range dirList {
		dirPath := p + "/" + dir.Name()
		prefixStr = getPrefix(dirPath, pf)
		dirStr := prefixStr + dir.Name()

		if dir.IsDir() {
			fmt.Fprintln(w, dirStr)
			dirTree(w, dirPath, pf)
		} else {
			if pf {
				fmt.Fprintln(w, dirStr, getFileSize(dir))
			}
		}
	}
	return nil
}

func getPrefix(fp string, pf bool) string {
	fpList := strings.Split(fp, "/")
	var preStr string
	dirPath := fpList[0]
	for i := 1; i < len(fpList); i++ {

		dirList, _ := ioutil.ReadDir(dirPath)
		if !pf {
			tmpList := dirList
			dirList = nil
			for _, dir := range tmpList {
				if dir.IsDir() {
					dirList = append(dirList, dir)
				}
			}
		}
		dirPath += "/" + fpList[i]
		//fmt.Println(fpList, i, len(fpList)-1)
		if i == len(fpList)-1 {
			if dirList != nil && fpList[i] == dirList[len(dirList)-1].Name() {
				preStr += "└───"
			} else {
				preStr += "├───"
			}
		} else {
			if fpList[i] == dirList[len(dirList)-1].Name() {
				preStr += "\t"
			} else {
				preStr += "│\t"
			}
		}
	}
	return preStr
}

func getFileSize(fi os.FileInfo) string {
	fiSize := fi.Size()
	if fiSize == 0 {
		return "(empty)"
	}

	return fmt.Sprintf("(%db)", fiSize)
}
