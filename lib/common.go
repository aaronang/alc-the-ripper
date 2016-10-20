package lib

import "encoding/json"

const (
	Protocol = "http://"
	Port     = ":8080"

	BodyType = "application/json"

	CreateTaskPath = "/tasks/create"
)

type Task struct {
	Id        int
	JobId     int
	Algorithm string
	Salt      string
	Digest    string
	CharSet   string
	Length    int
	Start     string
	End       string
}

func (t *Task) ToJson() ([]byte, error) {
	return json.Marshal(&t)
}
