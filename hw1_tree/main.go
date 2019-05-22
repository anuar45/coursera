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

func dirTree(w io.Writer, p string, files bool) error {
	var prefixStr string
	dirList, _ := ioutil.ReadDir(p)

	for i, dir := range dirList {
		dirPath := p + "/" + dir.Name()
		prefixStr := getPrefix(dirPath)
		dirStr := prefixStr + dir.Name()

		fmt.Fprintln(w, dirStr)

		dirTree(w, dirPath, files)
	}
	return nil
}

func getPrefix(fp string) string {
	fpList := strings.Split(fp, "/")
	//fmt.Println(fpList)
	var tabStr string
	dirPath := fpList[0]
	for i := 1; i < len(fpList); i++ {

		//fmt.Println(dirPath)
		dirList, _ := ioutil.ReadDir(dirPath)
		//fmt.Println(dirList)

		dirPath += "/" + fpList[i]
		//if dirList == nil {
		//	return tabStr + " \t"
		//}
		//fmt.Println(fpList[i], dirList[len(dirList)-1].Name())
		if i != len(dirList)-1 {
			prefixStr = "├───────"
		} else {
			prefixStr = "└───────"
		}
		if i != 1 {
			if fpList[i] == dirList[len(dirList)-1].Name() {
				tabStr += " \t"
			} else {
				tabStr += "│\t"
			}
		}
	}
	return tabStr
}
