package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Job struct {
	Salt    string
	Digest  string
	Length  int
	CharSet string
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&j)
	fmt.Println(j)
}

func main() {
	http.HandleFunc("/", jobsHandler)
	http.ListenAndServe(":8080", nil)
}
