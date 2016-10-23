package slave

type CrackerSuccess struct {
	TaskID   int
	Password string
}

type CrackerFail struct {
	TaskID   int
}
