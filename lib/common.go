package lib

import "encoding/json"

// Global configuration
const (
	Port     = ":8080"
	Protocol = "http://"

	BodyType = "application/json"

	CreateTaskPath = "/tasks/create"

	SlaveARN   = "arn:aws:iam::415077340068:instance-profile/SlaveTheRipper"
	SlaveRole  = "SlaveTheRipper"
	SlaveImage = "ami-7abd0209" // CentOS 7 (x86_64) - with Updates HVM for EU (Ireland)
	SlaveType  = "t2.micro"

	MasterARN  = "arn:aws:iam::415077340068:instance-profile/MasterTheRipper"
	MasterRole = "MasterTheRipper"

	AWSRegion = "eu-west-1"
)

// A Task defines the computational domain for string permutations. This way,
// the slave knows from which string permutation to start and at which string
// permutation to stop.
type Task struct {
	Job
	ID      int
	Start   []byte
	TaskLen int
}

// ToJSON serializes a Task to JSON.
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

type Algorithm int

const (
	PBKDF2 Algorithm = iota
	BCRYPT
	SCRYPT
	ARGON2
)

type CharSet int

const (
	Numerical CharSet = iota
	AlphaLower
	AlphaMixed
	AlphaNumLower
	AlphaNumMixed
)

// Job is the customer facing resource,
// we use naming conventions of https://godoc.org/golang.org/x/crypto/pbkdf2,
// and assuming the hash function is sha256
type Job struct {
	Salt      []byte
	Digest    []byte
	KeyLen    int
	Iter      int
	CharSet   CharSet
	Algorithm Algorithm
}
