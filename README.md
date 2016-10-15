# Alc the Ripper
Alc the Ripper (Alc for short) is a *state-of-the-art* cloud-based password cracker.
The user submits a salt, a PBKDF2 digest, and the length of the password, then Alc will use brute force to find the password.

# Design Requirements
* Automation - Alc should not require any human intervention when running. The user simply submits a request and Alc will do its job and then report the finding.
* Elasticity - Alc should handle variable user demands and offer the same level of service (in terms of hashes per second) to all the users.
* Load Balancing - The jobs should be evenly spread out across all the slaves to achieve maximum performance.
* Reliability - Jobs should checkpointed and restarted from the checkpoint when failures occur.
* Monitoring - Maintain metrics about the whole system to monitor job status, resource usage and so on.
* Scheduling - TODO
* Multi-tenancy - TODO
* Security - TODO

# Design Overview
Alc is designed to run on IaaS providers such as AWS.
It uses a master-slave model. Where the master assign tasks to the slaves and the slaves periodically send heartbeat messages back to the master.
The heartbeat messages may contain additional status infomation on the slave.

We achieve elasticity using the [PID controller](https://en.wikipedia.org/wiki/PID_controller).
Before getting into its functionality, we first define the meaning of resources in Alc.
Resources are tasks that can be carried out in parallel in a time instant.
For example, if there is an EC2 instance and it can run two tasks in parallel, then we have two units of resource.
For the PID controller, we define the error as (total available resource - total required resource).
The total required resource depends on the number of submitted tasks.
If the number of submitted tasks exceeds the maximum possible resource, e.g. 20 instances * 2 parallel tasks,
then the tasks need to be queued and the total required resource is capped at 20 * 2.
Only the proportional and the derivative terms are used in our controller.
The derivative term is to introduce damping, so that we don't spontaneous reserve and release instances.

Load balancing is achieved using the greedy load balancing algorithm where each new task is simply sent to the least loaded slave, or the slave that is expected to finish first if all slaves are busy.
This is possible because computing SHA2 is deterministic so we can estimate when every task will finish.
The greedy algorithm is not optimal under normal circumstances, e.g. it does a bad job if all the tasks are short but the final task is long.
However, in our case it is possible to fix the duration of the tasks by chunking them into subtasks and then combining the results after all the subtasks are completed.
For example a task might be - brute force a 6 character lowercase only password, then it is possible to split it into two subtasks where the first is responsible to compute all hashes beginning with a to m and the second is responsible for n-z.
If the duration of all subtasks are all the same or similar, we claim the greedy algorithm is optimal.

To achieve reliability on the master, we introduce a backup master. 
The master periodically synchronise its internal data structure with the backup, and the backup is responsible to detect when the master is offline and then take over if necessary.
It is possible to have multiple backup masters. The leader can be elected using the [Bully algorithm](https://en.wikipedia.org/wiki/Bully_algorithm).
The reliability on the slaves is achieved via the master.
The master internally stores the task status and slave status.
If the slave misses too many heartbeat messages then the master can assume it crashed and it will create a new slave by provision a new instance.

