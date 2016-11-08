package main

import (
	"log"

	"github.com/aaronang/cong-the-ripper/lib/master"
	"github.com/ogier/pflag"
)

func main() {
	log.Println("master starting...")
	port := pflag.String("port", "8080", "Web server port")
	ip := pflag.String("ip", "localhost", "Master IP")
	kp := pflag.Float64("kp", 1, "Proportional gain")
	ki := pflag.Float64("ki", 0, "Integral gain")
	kd := pflag.Float64("kd", 0, "Differential gain")
	pflag.Parse()

	m := master.Init(*port, *ip, *kp, *ki, *kd)
	m.Run()
}
