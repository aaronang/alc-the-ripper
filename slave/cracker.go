package slave

import (
	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/slave/brutedict"
	"github.com/aaronang/cong-the-ripper/slave/hasher"
)

func (slave *Slave) Execute(task lib.Task) {
	bd := brutedict.New(&task)
	var hasher hasher.Hasher = hasher.Pbkdf2{} // Can be swapped with other hashing algorithms

	for {
		if candidate := bd.Next(); candidate != nil {
			hash := hasher.Hash(candidate, &task)
			// fmt.Println("Key base64: " + string(candidate) + " -> " + b64.StdEncoding.EncodeToString(hash))
			if lib.TestEqByteArray(hash, task.Digest) {
				slave.successChan <- CrackerSuccess{taskID: task.ID, password: string(candidate)}
				break
			}
		} else {
			slave.failChan <- CrackerFail{taskID: task.ID}
			bd.Close()
			break
		}
	}
}
