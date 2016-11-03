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
	pflag.Parse()

	m := master.Init(*port, *ip)
	m.Run()
}
