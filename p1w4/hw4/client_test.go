package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Persons []Person

type Root struct {
	XMLName xml.Name `xml:root`
	List    []Person `xml:"row"`
	Version string   `"xml:"version,attr"`
}

const (
	validToken   = "ValidToken"
	invalidToken = "InvalidToken"
)

type Person struct {
	ID            int    `json:"Id"            xml:"id"`
	Giud          string `json:"guid"          xml:"guid"`
	IsActive      bool   `json:"isActive"      xml:"isActive"`
	Balance       string `json:"balance"       xml:"balance"`
	Picture       string `json:"picture"       xml:"picture"`
	Age           int    `json:"Age"           xml:"age"`
	EyeColor      string `json:"eyeColor"      xml:"eyeColor"`
	FirstName     string `json:"first_name"    xml:"first_name"`
	LastName      string `json:"last_name"     xml:"last_name"`
	Male          string `json:"Gender"        xml:"gender"`
	Company       string `json:"company"       xml:"company"`
	Email         string `json:"email"         xml:"email"`
	Phone         string `json:"phone"         xml:"phone"`
	Address       string `json:"address"       xml:"address"`
	About         string `json:"About"         xml:"about"`
	Registered    string `json:"registered"    xml:"registered"`
	FavoriteFruit string `json:"favoriteFruit" xml:"favoriteFruit"`
	Name          string `json:"Name"`
}

type PersonSorter interface {
	SortById()
	SortByName()
	SortByAge()
}

type PersonFilter interface {
	Filter()
}

type BadErrorResponse struct {
	Error int
}

// TODO: It is a handler for httptest server, should return data from dataset.xml
func SearchServer(w http.ResponseWriter, r *http.Request) {
	we := json.NewEncoder(w)
	q := r.FormValue("query")
	sortKey := r.FormValue("order_field")
	orderBy := r.FormValue("order_by")
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	token := r.Header.Get("AccessToken")
	if token == "" || token != validToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	xmlPath := "./dataset.xml"
	persons := GetPersons(xmlPath)

	//fmt.Println(persons)

	persons = persons.Filter(q)

	switch sortKey {
	case "":
		persons.SortByName(orderBy)
	case "Id":
		persons.SortByID(orderBy)
	case "Age":
		persons.SortByAge(orderBy)
	case "Name":
		persons.SortByName(orderBy)
	default:
		w.WriteHeader(http.StatusBadRequest)
		we.Encode(SearchErrorResponse{Error: "ErrorBadOrderField"})
		return
	}

	if offset > len(persons) {
		w.WriteHeader(http.StatusBadRequest)
		we.Encode(SearchErrorResponse{Error: "ErrorBadOffsetField"})
		return
	}

	if offset+limit > len(persons) {
		limit = len(persons) - offset
	}

	persons = persons[offset : offset+limit]

	//fmt.Println(offset, limit)
	//for _, p := range persons {
	//	fmt.Print(p.Name)
	//}

	w.WriteHeader(http.StatusOK)
	we.Encode(persons)
}

func SearchServerTimeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(101 * time.Second)
	w.WriteHeader(http.StatusOK)
}

func SearchServerFatal(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func SearchServerBadResponse(w http.ResponseWriter, r *http.Request) {
	we := json.NewEncoder(w)
	w.WriteHeader(http.StatusBadRequest)
	we.Encode(BadErrorResponse{Error: 9999})
}

func SearchServerErrorUnmarshal(w http.ResponseWriter, r *http.Request) {
	we := json.NewEncoder(w)
	w.WriteHeader(http.StatusOK)
	we.Encode(struct {
		Nickname string
	}{
		Nickname: "Johnny",
	})
}

func GetPersons(fp string) Persons {
	var r Root
	f, err := os.Open(fp)
	if err != nil {
		log.Fatalf("Cant open file %s:", err)
	}

	data, err := ioutil.ReadAll(f)
	err = xml.Unmarshal(data, &r)
	if err != nil {
		fmt.Println(err)
	}

	persons := r.List

	for _, p := range persons {
		p.Name = p.FirstName + p.LastName
	}

	return persons
}

func TestGetPersons(t *testing.T) {
	xmlPath := "./dataset.xml"
	GetPersons(xmlPath)
	//fmt.Println(ps)
}

func (persons Persons) Filter(q string) Persons {
	var personsMatched Persons

	if q == "" {
		return persons
	}

	for _, p := range persons {
		if strings.Contains(p.Name, q) || strings.Contains(p.About, q) {
			personsMatched = append(personsMatched, p)
		}
	}
	return personsMatched
}

func (persons Persons) SortByID(orderBy string) {
	switch orderBy {
	case "1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].ID < persons[j].ID })
	case "-1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].ID > persons[j].ID })
	case "0":
		return
	}
}

func (persons Persons) SortByName(orderBy string) {
	switch orderBy {
	case "1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].Name < persons[j].Name })
	case "-1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].Name > persons[j].Name })
	case "0":
		return
	}
}

func (persons Persons) SortByAge(orderBy string) {
	switch orderBy {
	case "1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].Age < persons[j].Age })
	case "-1":
		sort.Slice(persons, func(i, j int) bool { return persons[i].Age > persons[j].Age })
	case "0":
		return
	}
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
		Offset:     0,
		Query:      "Annie",
		OrderField: "Name",
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
		Offset:     0,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}
	srLimitMax := SearchRequest{
		Limit:      27,
		Offset:     0,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}
	srOffsetMin := SearchRequest{
		Limit:      5,
		Offset:     -2,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}
	srAll := SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "",
		OrderField: "Name",
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

	resp, err = sc.FindUsers(srAll)
	if !resp.NextPage {
		t.Errorf("Expected NextPage to be true Got: %v", resp)
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
		OrderField: "Name",
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
		OrderField: "Name",
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
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := validToken

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     10,
		Query:      "Annie",
		OrderField: "Password",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return \n Got: %v", resp)
	}
}

func TestFindUsersBadResponseErrorUnmarshal(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadResponse))
	token := validToken

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     10,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return Bad request and error json\n Got: %v", resp)
	}
}

func TestFindUsersUnknownError(t *testing.T) {
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
		Offset:     10,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return unknown error but got: %v", resp)
	}
}

func TestFindUsersUnmarshalError(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrorUnmarshal))
	token := validToken

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     10,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return error umarshaling json\n Got: %v", resp)
	}
}

func TestFindUsersNetError(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	//ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	token := validToken

	sc := SearchClient{
		AccessToken: token,
		URL:         "https://x8ejdmwehfkscm2.test",
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return network error, but got some result: %v", resp)
	}
}

func TestFindUsersTimeoutError(t *testing.T) {
	// Here you should instatiate your test server with your handler
	// and pass url of test server to call FindUsers
	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeout))
	token := validToken

	sc := SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}

	sr := SearchRequest{
		Limit:      5,
		Offset:     10,
		Query:      "Annie",
		OrderField: "Name",
		OrderBy:    1,
	}

	resp, err := sc.FindUsers(sr)
	if err == nil {
		t.Errorf("Should return timeout error\n Got: %v", resp)
	}
}
