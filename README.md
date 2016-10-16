# Alc the Ripper
Alc the Ripper (Alc for short) is a *state-of-the-art* cloud-based password cracker.
The user submits a salt, a PBKDF2 digest, and the length of the password, then Alc will use brute force to find the password.

# Design Requirements
* Automation - Alc should not require any human intervention when running. The user simply submits a request and Alc will do its job and then report the finding.
* Elasticity - Alc should handle variable user demands and offer the same level of service (in terms of hashes per second) to all the users.
* Load Balancing - The jobs should be evenly spread out across all the slaves to achieve maximum performance.
* Reliability - Jobs should checkpointed and restarted from the checkpoint when failures occur. A checkpoint is an intermediate state for which a job can start.
* Monitoring - Maintain metrics about the whole system to monitor job status, resource usage and so on.
* Scheduling - TODO
* Multi-tenancy - TODO
* Security - TODO

# Design Overview
Alc is designed to run on IaaS providers such as AWS.
It uses the master-slave model, where the master assigns tasks to the slaves and the slaves periodically send heartbeat messages and status updates back to the master.

We achieve elasticity using the [PID controller](https://en.wikipedia.org/wiki/PID_controller).
Before getting into its functionality, we first define the meaning of resources in Alc.
The number of resources is the number of tasks that can be carried out in parallel in a time instant, somewhat analogous to [DOP](https://en.wikipedia.org/wiki/Degree_of_parallelism).
For example, if an EC2 instance can run two tasks in parallel, then that instance represents two units of resource.
For the PID controller, we define the error as `Ra - Rr`, where `Ra` is the currently available resource, and `Rr` is the total required resource.
`Rr` depends on the number of submitted tasks and `Ra` is total number of instances that Alc provisioned multiplied by the DOP of every instance.
If the number of submitted tasks exceeds the maximum possible resource, e.g. 20 instances each with DOP of 2,
then the tasks need to be queued and the `Rr` is capped at 20 * 2.
In other words, queues should be not affect `Rr`, but they should never build up if we have enough resources.
The queuing happens on the master.
Only the proportional and the derivative terms are used in our controller.
The derivative term is to introduce damping, so that we don't spontaneous reserve and release instances.

Load balancing is achieved using the greedy load balancing algorithm where each new task is simply sent to the least loaded slave (in terms of the number of free "slots" it has available).
Queues should not build up in the slaves, but it is OK to build up on the master when Alc cannot scale any further.
Since the SHA2 computation is deterministic, it is also possible to estimate the finish time of tasks and push new tasks to the slaves just before the tasks are finished to minimise waiting time on the slave.
The greedy algorithm is not optimal under normal circumstances, e.g. it does a bad job if all the tasks are short but the final task is long.
However, in our case it is possible to fix the duration of the tasks by chunking them into sub-tasks and then combining the results after all the sub-tasks are completed.
For example a task might be - brute force a 6 character lowercase only password, then it is possible to split it into two sub-tasks where the first is responsible to compute all hashes beginning with a-m and the second is responsible for n-z.
If the duration of all sub-tasks are all the same or similar, we claim the greedy algorithm is optimal.
We set a global sub-task size in terms of the number of SHA2 hashes.

To achieve reliability on the master, we introduce a backup master. 
The master periodically synchronises its internal data structure with the backup, and the backup is responsible for detecting when the master is offline and then take over if necessary.
It is possible to have multiple backup masters. The master (leader) can be elected using the [Bully algorithm](https://en.wikipedia.org/wiki/Bully_algorithm) where new leaders are automatically elected when the current leader goes offline.
The reliability on the slaves is achieved via the master.
The master internally stores the task status and slave status, effectively creating checkpoints for every task.
If the slave misses too many heartbeat messages then the master can assume it crashed and it will create a new slave by provisioning a new instance and start the task from the last known checkpoint.

