package main

import (
	"log"

	"github.com/aaronang/cong-the-ripper/lib/slave"
	flag "github.com/ogier/pflag"
)

func main() {
	log.Println("slave starting...")
	port := flag.String("port", "8080", "Web server port")
	masterIP := flag.String("master-ip", "localhost", "Ip address of the master")
	masterPort := flag.String("master-port", "8080", "Port of the master")
	flag.Parse()

	s := slave.Init("instance.EC2.cong1", *port, *masterIP, *masterPort)
	s.Run()
}
