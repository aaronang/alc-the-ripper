package master

import (
	"bytes"
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
)

func TestNextCandidate(t *testing.T) {
	res1, overflow1 := nthCandidateFrom(lib.AlphaLower, 1, []byte("aaaa"))
	if bytes.Compare(res1, []byte("baaa")) != 0 {
		t.Error("alpha next candidate failed (1)")
	}
	if overflow1 {
		t.Error("alpha next candidate overflow not zero (1)")
	}

	_, overflow2 := nthCandidateFrom(lib.AlphaLower, 2, []byte("yzzz"))
	if !overflow2 {
		t.Error("alpha next candidate should overflow (2)")
	}

	res3, overflow3 := nthCandidateFrom(lib.AlphaLower, 1, []byte("zaaa"))
	if bytes.Compare(res3, []byte("abaa")) != 0 {
		t.Error("alpha next candidate failed (3)")
	}
	if overflow3 {
		t.Error("alpha next candidate overflow not zero (3)")
	}

	res4, overflow4 := nthCandidateFrom(lib.AlphaLower, 26*26, []byte("aaaa"))
	if bytes.Compare(res4, []byte("aaba")) != 0 {
		t.Error("alpha next candidate failed (4)")
	}
	if overflow4 {
		t.Error("alpha next candidate overflow not zero (4)")
	}
}

func TestChunkCandidates(t *testing.T) {
	combs, lens := chunkCandidates(lib.AlphaLower, 3, 26*26)
	if !testLastIsFinal(lib.AlphaLower, combs, lens) {
		t.Error("failed to chunk 3 alpha chars - last candidate is not final")
	}
	if len(combs) != 26 || len(lens) != 26 {
		t.Error("failed to chunk 3 alpha chars - wrong length")
	}

	combs2, lens2 := chunkCandidates(lib.Numerical, 2, 8)
	if !testLastIsFinal(lib.Numerical, combs2, lens2) {
		t.Error("failed to chunk 4 num chars - last candidate is not final")
	}
	if len(combs2) != 13 || len(lens2) != 13 {
		t.Error("failed to chunk 4 num chars - wrong length")
	}

	combs3, lens3 := chunkCandidates(lib.AlphaNumLower, 6, 1024*1024)
	if !testLastIsFinal(lib.AlphaNumLower, combs3, lens3) {
		t.Error("failed to chunk 6 alpha num chars - last candidate is not final")
	}
	if len(combs3) != 2076 || len(lens3) != 2076 {
		t.Error("failed to chunk 6 alpha num chars - wrong length")
	}
}

func testLastIsFinal(alph lib.Alphabet, combs [][]byte, lens []int64) bool {
	l := lens[len(lens)-1]
	b := combs[len(combs)-1]
	final := alph.FinalCandidate(len(b))
	final2, overflow := nthCandidateFrom(alph, l, b)
	if overflow || bytes.Compare(final2, final) != 0 {
		return false
	}
	return true
}
