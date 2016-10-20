package master

import "testing"
import "github.com/aaronang/cong-the-ripper/lib"
import "bytes"

func TestBytesToIntSlice(t *testing.T) {
	if !testEq(bytesToIntSlice(lib.Numerical, []byte("012309")), []int{0, 1, 2, 3, 0, 9}) {
		t.Error("numerical conversion failed")
	}

	if !testEq(bytesToIntSlice(lib.AlphaLower, []byte("abcxyz")), []int{0, 1, 2, 23, 24, 25}) {
		t.Error("alphanum conversion failed")
	}
}

func TestAddToIntArray(t *testing.T) {
	res1, rem1 := addToIntSlice(24, 32, []int{7, 21, 13})
	if !testEq(res1, []int{15, 22, 13}) {
		t.Error("base 24 add failed (1)")
	}

	if rem1 != 0 {
		t.Error("base 24 add remainder not 0 (1)")
	}

	res2, rem2 := addToIntSlice(24, 32, []int{20, 21, 13})
	if !testEq(res2, []int{4, 23, 13}) {
		t.Error("base 24 add failed (2)")
	}

	if rem2 != 0 {
		t.Error("base 24 add remainder not 0 (2)")
	}

	res3, rem3 := addToIntSlice(2, 1, []int{1, 1})
	if !testEq(res3, []int{0, 0}) {
		t.Error("binary add failed")
	}

	if rem3 != 1 {
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

func testEq(a, b []int) bool {

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
