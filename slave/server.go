package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/task"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var t task.Task
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&t)
	fmt.Println(t)
}

func main() {
	http.HandleFunc("/tasks/create", taskHandler)
	http.ListenAndServe(":8080", nil)
}
