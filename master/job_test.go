package master

import (
	"bytes"
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
)

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

func TestChunkCandidates(t *testing.T) {
	combs, lens := chunkCandidates(lib.AlphaLower, 3, 26*26)
	if len(combs) != 26 || len(lens) != 26 {
		t.Error("failed to chunk 3 alpha chars - wrong length")
	}

	combs2, lens2 := chunkCandidates(lib.Numerical, 2, 8)
	if len(combs2) != 13 || len(lens2) != 13 {
		t.Error("failed to chunk 4 num chars - wrong length")
	}

	combs3, lens3 := chunkCandidates(lib.AlphaNumLower, 6, 1024*1024)
	if len(combs3) != 2076 || len(lens3) != 2076 {
		t.Error("failed to chunk 6 alpha num - wrong length")
	}
}

/*
func testLastIsFinal(alph lib.Alphabet, combs [][]byte, lens []int) bool {
	l := lens[len(lens)-1]
	b := combs[len(combs)-1]
	final := alph.FinalCandidate(len(b))
	final2, carry := nthCandidateFrom(alph, l, b)
	if carry != 0 || bytes.Compare(final2, final) != 0 {
		return false
	}
	return true
}
*/
