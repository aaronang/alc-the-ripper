package main

import (
	"net/http"
)

type Subtask struct {
	Algorithm string
	Salt      string
	Digest    string
	CharSet   string
	Length    int
	Start     string
	End       string
	Id        int
}

func subtaskHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func main() {
	http.HandleFunc("/", subtaskHandler)
	http.ListenAndServe(":8080", nil)
}
