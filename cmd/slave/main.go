package main

import (
	"log"

	flag "github.com/ogier/pflag"

	"github.com/aaronang/cong-the-ripper/lib/slave"
)

func main() {
	log.Println("slave starting...")
	portPtr := flag.String("port", "8080", "Web server port")
	masterIpPtr := flag.String("master-ip", "localhost", "Ip address of the master")
	masterPortPtr := flag.String("master-port", "8080", "Port of the master")
	flag.Parse()

	s := slave.Init("instance.EC2.cong1", *portPtr, *masterIpPtr, *masterPortPtr)
	s.Run()
}
