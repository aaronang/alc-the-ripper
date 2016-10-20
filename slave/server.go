package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var t lib.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		return
	}
	fmt.Println(t)
}

func main() {
	http.HandleFunc(lib.CreateTaskRoute, taskHandler)
	http.ListenAndServe(lib.Port, nil)
}
