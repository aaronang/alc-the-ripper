package lib

import (
	"bytes"
	"encoding/json"
	"math/big"
)

// Global configuration
const (
	Port     = ":8080"
	Protocol = "http://"

	BodyType = "application/json"

	TasksCreatePath = "/tasks/create"
	JobsCreatePath  = "/jobs/create"
	HeartbeatPath   = "/heartbeat"
	StatusPath      = "/cong"

	SlaveARN   = "arn:aws:iam::415077340068:instance-profile/SlaveTheRipper"
	SlaveRole  = "SlaveTheRipper"
	SlaveImage = "ami-7abd0209" // CentOS 7 (x86_64) - with Updates HVM for EU (Ireland)
	SlaveType  = "t2.micro"

	MasterARN  = "arn:aws:iam::415077340068:instance-profile/MasterTheRipper"
	MasterRole = "MasterTheRipper"

	AWSRegion = "eu-west-1"
)

type Heartbeat struct {
	SlaveId    string // aws.Instance.InstanceId
	TaskStatus []TaskStatus
}

type TaskStatus struct {
	Id       int
	JobId    int
	Done     bool
	Password string
	Progress []byte // State of permutation
}

// A Task defines the computational domain for string permutations. This way,
// the slave knows from which string permutation to start and at which string
// permutation to stop.
type Task struct {
	Job
	JobID   int
	ID      int
	Start   []byte
	TaskLen int64
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

func (alph Alphabet) replicateAt(l, idx int) []byte {
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
	Digest    string
	KeyLen    int
	Iter      int
	Alphabet  Alphabet
	Algorithm Algorithm
}

// Alphabets contains the set of all candidate characters for every alphabet
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

// BytesToIntSlice is DEPRECATED
func BytesToIntSlice(alph Alphabet, inp []byte) []int {
	res := make([]int, len(inp))
	for i, b := range inp {
		// probably not efficient, but the character sets are small so it's negligible
		x := bytes.IndexByte(Alphabets[alph], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	return res
}

// AddToIntSlice is DEPRECATED
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

// IntSliceToBytes is DEPRECATED
func IntSliceToBytes(alph Alphabet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = Alphabets[alph][v]
	}
	return res
}

func BytesToBigInt(alph Alphabet, inp []byte) *big.Int {
	base := big.NewInt(int64(len(Alphabets[alph])))
	res := big.NewInt(0)
	for i, b := range inp {
		// probably not efficient, but the alphabets are small so it's negligible
		x := bytes.IndexByte(Alphabets[alph], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		z := big.NewInt(0)
		z.Exp(base, big.NewInt(int64(i)), nil)
		z.Mul(z, big.NewInt(int64(x)))
		res.Add(res, z)
	}
	return res
}

// note that the input is consumed
func BigIntToBytes(alph Alphabet, bigInt *big.Int, l int) ([]byte, bool) {
	x := big.NewInt(0).Set(bigInt)
	base := big.NewInt(int64(len(Alphabets[alph])))
	m := big.NewInt(0)
	zero := big.NewInt(0)
	var res []byte

	// we use the algorith described here
	// https://math.stackexchange.com/questions/111150/changing-a-number-between-arbitrary-bases
	for {
		x, m = x.DivMod(x, base, m)
		if len(m.Bytes()) == 0 {
			res = append(res, 0)
		} else {
			res = append(res, m.Bytes()...)
		}
		if x.Cmp(zero) == 0 {
			break
		}
	}

	// the length must be the same as the original byte
	// so we add extra zeros to the end if necessary
	overflow := false
	if len(res) < l {
		res = append(res, make([]byte, l-len(res))...)
	} else if len(res) > l {
		overflow = true
		res = res[:l]
	}

	// convert the byte slice to string readable
	for i := range res {
		res[i] = Alphabets[alph][res[i]]
	}

	return res, overflow
}

func ReverseArray(arr []byte) []byte {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
	return arr
}

// TestEqInts tests the equality between two int slices
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

func TestEqByteArray(a, b []byte) bool {

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
