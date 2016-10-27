package main

import (
	"flag"
	"fmt"

	"github.com/aaronang/cong-the-ripper/master"
)

func main() {
	fmt.Println("master starting...")
	portPtr := flag.String("port", "8080", "Web server port")
	flag.Parse()

	m := master.Init(*portPtr)
	m.Run()
}
