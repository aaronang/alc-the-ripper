package main

import (
	"flag"
	"fmt"

	"github.com/aaronang/cong-the-ripper/lib/slave"
)

func main() {
	fmt.Println("slave starting...")
	portPtr := flag.String("port", "8080", "Web server port")
	flag.Parse()

	s := slave.Init("instance.EC2.cong1", *portPtr)
	s.Run()
}
