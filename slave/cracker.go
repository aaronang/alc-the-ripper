package slave

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/slave/brutedict"
)

func Execute(task lib.Task, slave *Slave) {
	bd := brutedict.New(&task)
	digest, _ := b64.StdEncoding.DecodeString(task.Digest)

	for {
		if candidate := bd.Next(); candidate != nil {
			key := pbkdf2.Key(candidate, task.Salt, task.Iter, sha256.Size, sha256.New)
			fmt.Println("Key base64: " + string(candidate) + " -> " + b64.StdEncoding.EncodeToString(key))
			if lib.TestEqByteArray(key, digest) {
				slave.successChan <- CrackerSuccess{TaskID: task.ID, Password: string(candidate)}
				break
			}
		} else {
			slave.failChan <- CrackerFail{TaskID: task.ID}
			bd.Close()
			break
		}
	}
}
