package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Person ...
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

// Address ...
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

func getPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(&people)
}

func getPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

func createPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	person.ID = fmt.Sprintf("%d", len(people)+1)
	people = append(people, person)
	json.NewEncoder(w).Encode(person)
}

func deletePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
		}
	}
	json.NewEncoder(w).Encode(people)
}

func main() {
	println("Recoding the REST API in 5 minutes")
	router := mux.NewRouter()

	people = append(people, Person{ID: "1", Firstname: "Alex", Lastname: "Lee", Address: &Address{City: "Ho Chi Minh", State: "Tan Phu"}})
	people = append(people, Person{ID: "2", Firstname: "Minh", Lastname: "Le"})

	router.HandleFunc("/people", getPeopleEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", getPersonEndpoint).Methods("GET")
	router.HandleFunc("/people/add", createPersonEndpoint).Methods("POST")
	router.HandleFunc("/people/{id}", deletePersonEndpoint).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8888", router))
}

/*
Get people list:
	GET http://localhost:8888/people

Get person:
	GET http://localhost:8888/people/1

Create person:
	POST http://localhost:8888/people/add

	JSON Body
	{
		"firstname": "Hung",
		"lastname": "Tran",
		"address": {
			"city": "Seatle",
			"state": "Silicon Valey"
		}
	}

Detelet DELETE http://localhost:8888/people/3
*/
