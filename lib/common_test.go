package lib

import (
	"bytes"
	"testing"
)

func TestBytesToIntSlice(t *testing.T) {
	if !TestEqInts(BytesToIntSlice(Numerical, []byte("012309")), []int{0, 1, 2, 3, 0, 9}) {
		t.Error("numerical conversion failed")
	}

	if !TestEqInts(BytesToIntSlice(AlphaLower, []byte("abcxyz")), []int{0, 1, 2, 23, 24, 25}) {
		t.Error("alphanum conversion failed")
	}
}

func TestAddToIntArray(t *testing.T) {
	res1, carry1 := AddToIntSlice(24, 32, []int{7, 21, 13})
	if !TestEqInts(res1, []int{15, 22, 13}) {
		t.Error("base 24 add failed (1)")
	}

	if carry1 != 0 {
		t.Error("base 24 add remainder not 0 (1)")
	}

	res2, carry2 := AddToIntSlice(24, 32, []int{20, 21, 13})
	if !TestEqInts(res2, []int{4, 23, 13}) {
		t.Error("base 24 add failed (2)")
	}

	if carry2 != 0 {
		t.Error("base 24 add remainder not 0 (2)")
	}

	res3, carry3 := AddToIntSlice(2, 1, []int{1, 1})
	if !TestEqInts(res3, []int{0, 0}) {
		t.Error("binary add failed")
	}

	if carry3 != 1 {
		t.Error("binary add remainder not 1")
	}
}

func TestIntSliceToBytes(t *testing.T) {
	// convert from byte slice to int slice and then back should match the initial byte slice
	// TODO we need QuickCheck for these
	sliceNum := []byte("0690385669")
	if bytes.Compare(sliceNum, IntSliceToBytes(Numerical, BytesToIntSlice(Numerical, sliceNum))) != 0 {
		t.Error("numerical comparison failed")
	}

	sliceAlphaLower := []byte("zsdlfkjasreituxnkfzvlksd")
	if bytes.Compare(sliceAlphaLower, IntSliceToBytes(AlphaLower, BytesToIntSlice(AlphaLower, sliceAlphaLower))) != 0 {
		t.Error("alpha comparison failed")
	}

	sliceAlphaNumLower := []byte("z1dlf9kjasrei12xnzvk7sd0")
	if bytes.Compare(sliceAlphaNumLower, IntSliceToBytes(AlphaNumLower, BytesToIntSlice(AlphaNumLower, sliceAlphaNumLower))) != 0 {
		t.Error("alpha num comparison failed")
	}

	sliceAlphaMixed := []byte("UdlkfSDFHsdflFdZFg")
	if bytes.Compare(sliceAlphaMixed, IntSliceToBytes(AlphaMixed, BytesToIntSlice(AlphaMixed, sliceAlphaMixed))) != 0 {
		t.Error("alpha mixed comparison failed")
	}
}

func TestInitialCandidate(t *testing.T) {
	if bytes.Compare(AlphaLower.InitialCandidate(5), []byte("aaaaa")) != 0 {
		t.Error("failed to generate initial alpha")
	}

	if bytes.Compare(AlphaNumLower.InitialCandidate(4), []byte("0000")) != 0 {
		t.Error("failed to generate initial alpha")
	}
}

func TestFinalCandidate(t *testing.T) {
	if bytes.Compare(AlphaLower.FinalCandidate(5), []byte("zzzzz")) != 0 {
		t.Error("failed to generate final alpha")
	}

	if bytes.Compare(AlphaNumLower.FinalCandidate(4), []byte("zzzz")) != 0 {
		t.Error("failed to generate final alpha")
	}
}
