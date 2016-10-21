package master

import (
	"bytes"
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
)

func TestBytesToIntSlice(t *testing.T) {
	if !testEq(bytesToIntSlice(lib.Numerical, []byte("012309")), []int{0, 1, 2, 3, 0, 9}) {
		t.Error("numerical conversion failed")
	}

	if !testEq(bytesToIntSlice(lib.AlphaLower, []byte("abcxyz")), []int{0, 1, 2, 23, 24, 25}) {
		t.Error("alphanum conversion failed")
	}
}

func TestAddToIntArray(t *testing.T) {
	res1, carry1 := addToIntSlice(24, 32, []int{7, 21, 13})
	if !testEq(res1, []int{15, 22, 13}) {
		t.Error("base 24 add failed (1)")
	}

	if carry1 != 0 {
		t.Error("base 24 add remainder not 0 (1)")
	}

	res2, carry2 := addToIntSlice(24, 32, []int{20, 21, 13})
	if !testEq(res2, []int{4, 23, 13}) {
		t.Error("base 24 add failed (2)")
	}

	if carry2 != 0 {
		t.Error("base 24 add remainder not 0 (2)")
	}

	res3, carry3 := addToIntSlice(2, 1, []int{1, 1})
	if !testEq(res3, []int{0, 0}) {
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
	if bytes.Compare(sliceNum, intSliceToBytes(lib.Numerical, bytesToIntSlice(lib.Numerical, sliceNum))) != 0 {
		t.Error("numerical comparison failed")
	}

	sliceAlphaLower := []byte("zsdlfkjasreituxnkfzvlksd")
	if bytes.Compare(sliceAlphaLower, intSliceToBytes(lib.AlphaLower, bytesToIntSlice(lib.AlphaLower, sliceAlphaLower))) != 0 {
		t.Error("alpha comparison failed")
	}

	sliceAlphaNumLower := []byte("z1dlf9kjasrei12xnzvk7sd0")
	if bytes.Compare(sliceAlphaNumLower, intSliceToBytes(lib.AlphaNumLower, bytesToIntSlice(lib.AlphaNumLower, sliceAlphaNumLower))) != 0 {
		t.Error("alpha num comparison failed")
	}

	sliceAlphaMixed := []byte("UdlkfSDFHsdflFdZFg")
	if bytes.Compare(sliceAlphaMixed, intSliceToBytes(lib.AlphaMixed, bytesToIntSlice(lib.AlphaMixed, sliceAlphaMixed))) != 0 {
		t.Error("alpha mixed comparison failed")
	}
}

func TestNextCandidate(t *testing.T) {
	res1, carry1 := nthCandidateFrom(lib.AlphaLower, 1, []byte("aaaa"))
	if bytes.Compare(res1, []byte("baaa")) != 0 {
		t.Error("alpha next combination failed (1)")
	}
	if carry1 != 0 {
		t.Error("alpha next combination carry not zero (1)")
	}

	// the result will cycle back and result in a carry
	res2, carry2 := nthCandidateFrom(lib.AlphaLower, 2, []byte("yzzz"))
	if bytes.Compare(res2, []byte("aaaa")) != 0 {
		t.Error("alpha next combination failed (2)")
	}
	if carry2 != 1 {
		t.Error("alpha next combination carry not one (2)")
	}

	res3, carry3 := nthCandidateFrom(lib.AlphaLower, 1, []byte("zaaa"))
	if bytes.Compare(res3, []byte("abaa")) != 0 {
		t.Error("alpha next combination failed (3)")
	}
	if carry3 != 0 {
		t.Error("alpha next combination carry not zero (3)")
	}

	res4, carry4 := nthCandidateFrom(lib.AlphaLower, 26*26, []byte("aaaa"))
	if bytes.Compare(res4, []byte("aaba")) != 0 {
		t.Error("alpha next combination failed (4)")
	}
	if carry4 != 0 {
		t.Error("alpha next combination carry not zero (4)")
	}
}

func TestInitialCandidate(t *testing.T) {
	if bytes.Compare(initialCandidate(lib.AlphaLower, 5), []byte("aaaaa")) != 0 {
		t.Error("failed to generate initial alpha")
	}

	if bytes.Compare(initialCandidate(lib.AlphaNumLower, 4), []byte("0000")) != 0 {
		t.Error("failed to generate initial alpha")
	}
}

func TestChunkCharSet(t *testing.T) {
	combs, lens := chunkCharSet(lib.AlphaLower, 3, 26*26)
	if len(combs) != 26 || len(lens) != 26 {
		t.Error("failed to chunk 3 alpha chars - wrong length")
	}

	if !testLastIsFinal(lib.AlphaLower, combs, lens) {
		t.Error("failed to chunk 3 alpha chars - last combination is not final")
	}

	combs2, lens2 := chunkCharSet(lib.Numerical, 2, 8)
	if !testLastIsFinal(lib.Numerical, combs2, lens2) {
		t.Error("failed to chunk 4 num chars - last combination is not final")
	}

	combs3, lens3 := chunkCharSet(lib.AlphaNumLower, 6, 1024*1024)
	if !testLastIsFinal(lib.AlphaNumLower, combs3, lens3) {
		t.Error("failed to chunk 6 alpha num - last combination is not final")
	}
}

func testLastIsFinal(charset lib.CharSet, combs [][]byte, lens []int) bool {
	l := lens[len(lens)-1]
	b := combs[len(combs)-1]
	final := finalCandidate(charset, len(b))
	final2, carry := nthCandidateFrom(charset, l, b)
	if carry != 0 || bytes.Compare(final2, final) != 0 {
		return false
	}
	return true
}
