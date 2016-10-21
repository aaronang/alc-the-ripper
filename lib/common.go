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

// Job is the customer facing resource representing a single password cracking job
// we focus on PBKDF2 first with SHA256 first
type Job struct {
	Salt      []byte
	Digest    []byte
	KeyLen    int
	Iter      int
	CharSet   CharSet
	Algorithm Algorithm
}

// CharSetSlice contains the set of all candidate characters for every alphabet
var CharSetSlice [][]byte

func init() {
	const nums string = "0123456789"
	const alphaLower string = "abcdefghijklmnopqrstuvwxyz"
	const alphaMixed string = alphaLower + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const alphaNumLower string = nums + alphaLower
	const alphaNumMixed string = nums + alphaMixed
	CharSetSlice = [][]byte{
		[]byte(nums),
		[]byte(alphaLower),
		[]byte(alphaMixed),
		[]byte(alphaNumLower),
		[]byte(alphaNumMixed)}
}

func BytesToIntSlice(charset CharSet, inp []byte) []int {
	res := make([]int, len(inp))
	for i, b := range inp {
		// probably not efficient, but the character sets are small so it's negligible
		x := bytes.IndexByte(CharSetSlice[charset], b)
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

func IntSliceToBytes(charset CharSet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = CharSetSlice[charset][v]
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
