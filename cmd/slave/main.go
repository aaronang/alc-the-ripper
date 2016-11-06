package main

import (
	"log"

	"github.com/aaronang/cong-the-ripper/lib/slave"
	"github.com/ogier/pflag"
)

func main() {
	log.Println("slave starting...")
	port := pflag.String("port", "8080", "Web server port")
	masterIP := pflag.String("master-ip", "localhost", "Ip address of the master")
	masterPort := pflag.String("master-port", "8080", "Port of the master")
	pflag.Parse()

	s := slave.Init(*port, *masterIP, *masterPort)
	s.Run()
}
