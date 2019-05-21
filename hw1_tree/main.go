package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	//"path/filepath"
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
	identCount := strings.Count(p, "//")
	var tabs, prefix string
	dirList, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}

	for i, dir := range dirList {
		tabs = ""
		for j := 0; j < identCount; j++ {
			tabs += "|\t "
		}

		if i != len(dirList)-1 {
			prefix = "├───"
		} else {
			prefix = "└───"
		}
		dirStr := tabs + prefix + dir.Name()

		fmt.Fprintln(w, dirStr)
		dirPath := p + "//" + dir.Name()
		dirTree(w, dirPath, files)
	}
	return nil
}

func getPrefix(fp string) string {
	fpList := filepath.SplitList(fp)

}
