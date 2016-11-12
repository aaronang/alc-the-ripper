package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/ogier/pflag"
)

func main() {
	ip := pflag.IP("ip", net.IPv4(127, 0, 0, 1), "Master IP")
	port := pflag.String("port", "8080", "Web server port")
	interval := pflag.Int("interval", 10, "Interval between metric collection creation in seconds")
	output := pflag.String("output", "/tmp/cong"+time.Now().String()+".json", "Output filename")
	pflag.Parse()

	f, err := os.Create(*output)
	if err != nil {
		log.Panicln("[main] Creating file failed.", err)
	}

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		f.WriteString("]\n")
		f.Close()
		os.Exit(0)
	}()

	addr := lib.Protocol + net.JoinHostPort(ip.String(), *port) + lib.StatusPath
	f.WriteString("[")
	f.Write(getStatus(addr))

	for {
		time.Sleep(time.Duration(*interval) * time.Second)
		f.WriteString("\n,")
		f.Write(getStatus(addr))
	}
}

func getStatus(addr string) []byte {
	resp, err := http.Get(addr)
	if err != nil {
		log.Panicln("[main] Getting report failed.", err)
	}

	report, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln("[main] Cannot read response.", err)
	}
	return report
}
