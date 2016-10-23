package slave

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/slave/cracker"
)

type Slave struct {
}

func Init() Slave {
	// TODO initialise Slave correctly
	return Slave{}
}

func (s *Slave) Run() {
	http.HandleFunc(lib.TasksCreatePath, taskHandler)
	http.ListenAndServe(lib.Port, nil)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var t lib.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(t)
	go cracker.Execute(t)
}
