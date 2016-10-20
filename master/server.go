package master

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
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
	if err := decoder.Decode(&j); err != nil {
		http.Error(w, "Status: Bad Request", http.StatusBadRequest)
		return
	}
	fmt.Println(j)
}

func main() {
	http.HandleFunc("/", jobsHandler)
	http.ListenAndServe(lib.Port, nil)
}
