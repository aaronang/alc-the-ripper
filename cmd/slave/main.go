package main

import "fmt"
import "github.com/aaronang/cong-the-ripper/slave"

func main() {
	fmt.Println("slave starting...")
	s := slave.Init("instance.EC2.cong1")
	s.Run()
}
