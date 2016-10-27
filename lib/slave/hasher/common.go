package hasher

import "github.com/aaronang/cong-the-ripper/lib"

type Hasher interface {
	Hash(candidate []byte, task *lib.Task) []byte
}
