package master

import "testing"
import "github.com/aaronang/cong-the-ripper/lib"

func TestBytesToIntArray(t *testing.T) {
	if !testEq(bytesToIntArray(lib.Numerical, []byte("012309")), []int{0, 1, 2, 3, 0, 9}) {
		t.Error("numerical conversion failed")
	}

	if !testEq(bytesToIntArray(lib.AlphaLower, []byte("abcxyz")), []int{0, 1, 2, 23, 24, 25}) {
		t.Error("alphanum conversion failed")
	}
}

func TestAddToIntArray(t *testing.T) {
	if !testEq(addToIntArray(24, 32, []int{7, 21, 13}), []int{15, 22, 13}) {
		t.Error("base 24 add failed 1")
	}

	if !testEq(addToIntArray(24, 32, []int{20, 21, 13}), []int{4, 23, 13}) {
		t.Error("base 24 add failed 1")
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
