package slave

import (
	"log"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/lib/slave/brutedict"
	"github.com/aaronang/cong-the-ripper/lib/slave/hasher"
)

func execute(task *task, successChan chan CrackerSuccess, failChan chan CrackerFail) {
	bd := brutedict.New(&task.Task)
	hasher := new(hasher.Pbkdf2) // Can be swapped with other hashing algorithms
	var candidate []byte

	log.Println("[ Task", task.ID, "]", "Start cracker.Execute")
outer:
	for {
		select {
		case c := <-task.progressChan:
			// Return reversed array to match reversed encoding in the master
			c <- lib.ReverseArray(candidate)
		case <-task.killChan:
			// job killed, another slave should have the result
			failChan <- CrackerFail{taskID: task.ID}
			break outer
		default:
			if candidate = bd.Next(); candidate != nil {
				hash := hasher.Hash(candidate, &task.Task)
				if lib.TestEqBytes(hash, task.Digest) {
					successChan <- CrackerSuccess{taskID: task.ID, password: string(candidate)}
					break outer
				}
			} else {
				failChan <- CrackerFail{taskID: task.ID}
				break outer
			}
		}
	}
	log.Println("[ Task", task.ID, "]", "Finished cracker.Execute")
}
