package main

import "fmt"
import "github.com/aaronang/cong-the-ripper/master"

func main() {
	fmt.Println("master starting...")
	m := master.Init()
	m.Run()
}
