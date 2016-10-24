package slave

type CrackerSuccess struct {
	taskID   int
	password string
}

type CrackerFail struct {
	taskID   int
}
