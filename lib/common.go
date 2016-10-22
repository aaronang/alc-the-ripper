package lib

import (
	"bytes"
	"encoding/json"
)

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
	JobID   int
	ID      int
	Start   []byte
	TaskLen int
}

// ToJSON serializes a Task to JSON.
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// Algorithm is an enum for the supported key derivation functions
type Algorithm int

const (
	PBKDF2 Algorithm = iota
	BCRYPT
	SCRYPT
	ARGON2
)

// Alphabet
type Alphabet int

const (
	Numerical Alphabet = iota
	AlphaLower
	AlphaMixed
	AlphaNumLower
	AlphaNumMixed
)

func (alph Alphabet) InitialCandidate(l int) []byte {
	return alph.replicateAt(l, 0)
}

func (alph Alphabet) FinalCandidate(l int) []byte {
	return alph.replicateAt(l, len(Alphabets[alph])-1)
}

func (alph Alphabet) replicateAt(l int, idx int) []byte {
	v := Alphabets[alph][idx]
	res := make([]byte, l)
	for i := range res {
		res[i] = v
	}
	return res
}

// Job is the customer facing resource representing a single password cracking job
// we focus on PBKDF2 first with SHA256 first
type Job struct {
	Salt      []byte
	Digest    []byte
	KeyLen    int
	Iter      int
	Alphabet  Alphabet
	Algorithm Algorithm
}

// CharSetSlice contains the set of all candidate characters for every alphabet
var Alphabets [][]byte

func init() {
	const nums string = "0123456789"
	const alphaLower string = "abcdefghijklmnopqrstuvwxyz"
	const alphaMixed string = alphaLower + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const alphaNumLower string = nums + alphaLower
	const alphaNumMixed string = nums + alphaMixed
	Alphabets = [][]byte{
		[]byte(nums),
		[]byte(alphaLower),
		[]byte(alphaMixed),
		[]byte(alphaNumLower),
		[]byte(alphaNumMixed)}
}

func BytesToIntSlice(charset Alphabet, inp []byte) []int {
	res := make([]int, len(inp))
	for i, b := range inp {
		// probably not efficient, but the character sets are small so it's negligible
		x := bytes.IndexByte(Alphabets[charset], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	return res
}

func AddToIntSlice(base, v int, inp []int) ([]int, int) {
	for i := range inp {
		r := v % base
		tmp := inp[i] + r
		if tmp < base {
			inp[i] = tmp
			v = v / base
		} else {
			inp[i] = tmp % base
			v = v/base + 1
		}
	}
	return inp, v // v is the carry
}

func IntSliceToBytes(charset Alphabet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = Alphabets[charset][v]
	}
	return res
}

func TestEqInts(a, b []int) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
