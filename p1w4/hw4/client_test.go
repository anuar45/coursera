package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type Users struct {
	List []UserXML `xml:"root"`
}

type UserXML struct {
	ID            int    `json:"id"            xml:"id"`
	Giud          string `json:"guid"          xml:"guid"`
	IsActive      bool   `json:"isActive"      xml:"isActive"`
	Balance       string `json:"balance"       xml:"balance"`
	Picture       string `json:"picture"       xml:"picture"`
	Age           int    `json:"age"           xml:"age"`
	EyeColor      string `json:"eyeColor"      xml:"eyeColor"`
	FirstName     string `json:"first_name"    xml:"first_name"`
	LastName      string `json:"last_name"     xml:"last_name"`
	Male          string `json:"gender"        xml:"gender"`
	Company       string `json:"company"       xml:"company"`
	Email         string `json:"email"         xml:"email"`
	Phone         string `json:"phone"         xml:"phone"`
	Address       string `json:"address"       xml:"address"`
	About         string `json:"about"         xml:"about"`
	Registered    string `json:"registered"    xml:"registered"`
	FavoriteFruit string `json:"favoriteFruit" xml:"favoriteFruit"`
	Name          string
}

// TODO: It is a handler for httptest server, should return data from dataset.xml
func SearchServer(w http.ResponseWriter, r *http.Request) {
	// r.FormValues()
	xmlPath := "./dataset.xml"
	users := GetUsers(xmlPath)

	w.WriteHeader(http.StatusOK)
	we := json.NewEncoder(w)
	we.Encode(users.List)
}

func TestFindUsers(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := "SomeToken"

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     7,
		Query:      "Annie",
		OrderField: "FirstName",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err != nil {
		log.Fatalf("Error getting users: %s", err)
	}

	fmt.Println(resp)
}

func GetUsers(p string) Users {
	var users Users
	f, err := os.Open(p)
	if err != nil {
		log.Fatalf("Cant open file %s:", err)
	}

	xmlDecoder := xml.NewDecoder(f)
	xmlDecoder.Decode(users)

	for _, user := range users.List {
		user.Name = user.FirstName + user.LastName
	}

	return users
}
