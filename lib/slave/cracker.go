package slave

import (
	"log"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/lib/slave/brutedict"
	"github.com/aaronang/cong-the-ripper/lib/slave/hasher"
)

func Execute(task *task, successChan chan CrackerSuccess, failChan chan CrackerFail) {
	bd := brutedict.New(&task.Task)
	var hasher hasher.Hasher = new(hasher.Pbkdf2) // Can be swapped with other hashing algorithms

	log.Println("[ Task", task.ID, "]", "Start cracker.Execute")
	for {
		if candidate := bd.Next(); candidate != nil {
			select {
			case c := <-task.progressChan:
				c <- string(candidate)
			default:
			}

			hash := hasher.Hash(candidate, &task.Task)
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
	log.Println("[ Task", task.ID, "]", "Finished cracker.Execute")
}
