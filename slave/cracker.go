package slave

import (
	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/slave/brutedict"
	"github.com/aaronang/cong-the-ripper/slave/hasher"
)

func Execute(task lib.Task, successChan chan CrackerSuccess, failChan chan CrackerFail) {
	bd := brutedict.New(&task)
	var hasher hasher.Hasher = new(hasher.Pbkdf2) // Can be swapped with other hashing algorithms

	for {
		if candidate := bd.Next(); candidate != nil {
			hash := hasher.Hash(candidate, &task)
			// fmt.Println("Key base64: " + string(candidate) + " -> " + b64.StdEncoding.EncodeToString(hash))
			if lib.TestEqBytes(hash, task.Digest) {
				successChan <- CrackerSuccess{taskID: task.ID, password: string(candidate)}
				break
			}
		} else {
			failChan <- CrackerFail{taskID: task.ID}
			bd.Close()
			break
		}
	}
}
