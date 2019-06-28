package main

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type Persons struct {
	List []Person `xml:"root"`
}

const (
	validToken   = "ValidToken"
	invalidToken = "InvalidToken"
)

type Person struct {
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
	q := r.FormValue("query")
	token := r.Header.Get("AccessToken")
	if token == "" || token != validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	xmlPath := "./dataset.xml"
	persons := GetPersons(xmlPath)

	foundPersons := SearchPersons(persons, q)
	w.WriteHeader(http.StatusOK)
	we := json.NewEncoder(w)
	we.Encode(foundPersons)
}

func SearchServerFatal(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func GetPersons(p string) []Person {
	var persons Persons
	f, err := os.Open(p)
	if err != nil {
		log.Fatalf("Cant open file %s:", err)
	}

	xmlDecoder := xml.NewDecoder(f)
	xmlDecoder.Decode(persons)

	for _, p := range persons.List {
		p.Name = p.FirstName + p.LastName
	}

	return persons.List
}

func SearchPersons(persons []Person, q string) []Person {
	var personsMatched []Person
	for _, p := range persons {
		if strings.Contains(p.Name, q) || strings.Contains(p.About, q) {
			personsMatched = append(personsMatched, p)
		}
	}
	return personsMatched
}

func SortPersons(persons []Person, key string) {
	var personsSorted []Person
	//sort.Slice
}

func TestFindUsersSuccess(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := validToken

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
	if err != nil || resp == nil {
		t.Errorf("Error searching users: %v", err)
	}
}

func TestFindUsersParams(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	var resp *SearchResponse
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := "ValidToken"

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	srLimitMin := SearchRequest{
		Limit:      -1,
		Offset:     7,
		Query:      "Annie",
		OrderField: "FirstName",
		OrderBy:    1,
	}
	srLimitMax := SearchRequest{
		Limit:      27,
		Offset:     7,
		Query:      "Annie",
		OrderField: "FirstName",
		OrderBy:    1,
	}
	srOffsetMin := SearchRequest{
		Limit:      5,
		Offset:     -2,
		Query:      "Annie",
		OrderField: "FirstName",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(srLimitMin)
	if err == nil {
		t.Errorf("Should return error on negative Limit param\n Got: %v", resp)
	}

	resp, err = sc.FindUsers(srLimitMax)
	if err != nil {
		t.Errorf("Should not error on Limit param bigger then 25\n Got: %v", resp)
	}

	resp, err = sc.FindUsers(srOffsetMin)
	if err == nil {
		t.Errorf("Should return error on negative Offset param\n Got: %v", resp)
	}
}

func TestFindUsersStatusUnauthorised(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := invalidToken

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
	if err.Error() != "Bad AccessToken" {
		t.Errorf("Should return error on incorrect AcessToken header\n Got: %v", resp)
	}
}

func TestFindUsersStatusInternalServerError(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServerFatal))
	token := validToken

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
	if err.Error() != "SearchServer fatal error" {
		t.Errorf("Should return internal fatal error on broken server\n Got: %v", resp)
	}
}

func TestFindUsersStatusBadRequest(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServerFatal))
	token := validToken

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
	if err.Error() != "SearchServer fatal error" {
		t.Errorf("Should return internal fatal error on broken server\n Got: %v", resp)
	}
}
