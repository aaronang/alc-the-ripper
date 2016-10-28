// functions and methods related to the autoscaler

package master

import (
	"log"
	"sort"

	"github.com/aaronang/cong-the-ripper/lib"
)

// runController runs one iteration
func (m *Master) runController() {
	err := float64(m.countRequiredSlots() - m.countTotalSlots())

	dt := m.controller.dt.Seconds()
	m.controller.integral = m.controller.integral + err*dt
	derivative := (err - m.controller.prevErr) / dt
	output := m.controller.kp*err +
		m.controller.ki*m.controller.integral +
		m.controller.kd*derivative
	m.controller.prevErr = err

	log.Printf("err: %v, output: %v\n", err, output)
	m.adjustInstanceCount(int(output))
}

func (m *Master) adjustInstanceCount(n int) {
	if n > 0 {
		go func() {
			_, err := createSlaves(m.svc, n)
			if err != nil {
				log.Println("Failed to create slaves", err)
			} else {
				log.Printf("%v instances created successfully\n", n)
				// no need to report back to the master loop
				// because it should start receiving heartbeat messages
			}
		}()
	} else {
		// negate n to represent the (positive) number of instances to kill
		// scale by the number of
		n = -n / lib.MaxSlotsPerInstance
		if n == 0 {
			log.Println("n is 0 in adjustInstanceCount")
			return
		}

		// kills n least loaded slaves, the killed slaves may have unfinished tasks
		// but the master should detect missing heartbeats and restart the tasks
		ips := sortInstancesByTaskCount(m.instances)[:n]
		go func() {
			_, err := terminateSlaves(m.svc, instancesFromIPs(m.svc, ips))
			if err != nil {
				log.Println("Failed to terminate slaves", err)
			} else {
				log.Printf("%v instances terminated successfully", n)
				// again, no need to report success/failure
				// because heartbeat messages will stop
			}
		}()
	}
}

func (m *Master) maxSlots() int {
	// TODO do we set the manually or it's a property of AWS?
	return 20 * lib.MaxSlotsPerInstance
}

func (m *Master) countRequiredSlots() int {
	cnt := len(m.scheduledTasks) + len(m.newTasks)
	if cnt > m.maxSlots() {
		return m.maxSlots()
	}
	return cnt
}

func (m *Master) countTotalSlots() int {
	cnt := 0
	for _, i := range m.instances {
		cnt += i.maxSlots
	}
	return cnt
}

func sortInstancesByTaskCount(instances map[string]slave) []string {
	pairs := make([]ipSlavePair, len(instances))
	var i int
	for k, v := range instances {
		pairs[i].slave = v
		pairs[i].ip = k
		i++
	}

	sort.Sort(byTaskCount(pairs))

	ips := make([]string, len(pairs))
	for i := range ips {
		ips[i] = pairs[i].ip
	}
	return ips
}

type ipSlavePair struct {
	ip    string
	slave slave
}

type byTaskCount []ipSlavePair

func (a byTaskCount) Len() int {
	return len(a)
}

func (a byTaskCount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byTaskCount) Less(i, j int) bool {
	return len(a[i].slave.tasks) < len(a[j].slave.tasks)
}
