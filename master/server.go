package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
)

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(j)
}

func main() {
	http.HandleFunc("/", jobsHandler)
	http.ListenAndServe(lib.Port, nil)
}
