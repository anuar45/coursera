package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Users []User

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"-"`
	Country  string   `json:"-"`
	Email    string   `json:"email"`
	Job      string   `json:"-"`
	Name     string   `json:"name"`
	Phone    string   `json:"-"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file) // Optimize
	if err != nil {
		panic(err)
	}

	reAt := regexp.MustCompile("@")
	reAndroid := regexp.MustCompile("Android")
	reMSIE := regexp.MustCompile("MSIE")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	var users Users

	lines := strings.Split(string(fileContents), "\n")

	for _, line := range lines {
		var user User
		err = user.UnmarshalJSON([]byte(line))
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if reAndroid.MatchString(browser) {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browser := range user.Browsers {
			if reMSIE.MatchString(browser) {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := reAt.ReplaceAllString(user.Email, " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
