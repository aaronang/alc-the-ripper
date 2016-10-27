package main

import (
	"flag"
	"fmt"

	"github.com/aaronang/cong-the-ripper/master"
)

func main() {
	fmt.Println("master starting...")
	port := flag.String("port", "8080", "Web server port")
	flag.Parse()

	m := master.Init(*port)
	m.Run()
}
