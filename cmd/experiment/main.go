package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	mrand "math/rand"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/ogier/pflag"
	"golang.org/x/crypto/pbkdf2"
)

const (
	KeyLen    = 4
	Alphabet  = lib.AlphaLower
	Algorithm = lib.PBKDF2
	Iter      = 1000
)

var url string

func main() {
	ip := pflag.String("ip", "localhost", "Master IP")
	port := pflag.String("port", "8080", "Web server port")
	jobs := pflag.Int("jobs", 2, "Number of jobs to create")
	interval := pflag.Int("interval", 1, "Interval between job creation in seconds")
	output := pflag.String("output", "/tmp/cong"+time.Now().String(), "Output filename")
	pflag.Parse()

	url = lib.Protocol + net.JoinHostPort(*ip, *port)

	go func() {
		for i := 0; i < *jobs; i++ {
			if _, err := createJob(); err != nil {
				log.Panicln("[createJob] Job wasn't created properly.", err)
			}
			log.Println("[createJob] Job was created successfully.")
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	f, err := os.Create(*output)
	if err != nil {
		log.Panicln("[main] Creating file failed.", err)
	}
	f.WriteString("[")

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		f.WriteString("\r]\n")
		f.Close()
		os.Exit(0)
	}()

	for {
		resp, err := http.Get(url + lib.StatusPath)
		if err != nil {
			log.Panicln("[main] Getting report failed.", err)
		}

		report, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panicln("[main] Cannot read response.", err)
		}
		f.Write(report)
		f.WriteString("\n,")
		log.Println(string(report))

		time.Sleep(1 * time.Second)
	}
}

func createJob() (*http.Response, error) {
	body := generateJobJSON()
	return http.Post(url+lib.JobsCreatePath, lib.BodyType, bytes.NewBuffer(body))
}

func generateJobJSON() []byte {
	salt := generateSalt()
	pass := generatePassword()
	dig := generateDigest(pass, salt)

	j := lib.Job{
		Salt:      salt,
		Digest:    dig,
		KeyLen:    KeyLen,
		Iter:      Iter,
		Alphabet:  Alphabet,
		Algorithm: Algorithm,
	}
	body, err := j.ToJSON()
	if err != nil {
		log.Panicln(err)
	}
	return body
}

func generateSalt() []byte {
	b := make([]byte, mrand.Intn(6)+1)
	_, err := rand.Read(b)
	if err != nil {
		log.Panicln("error:", err)
	}
	return b
}

func generateDigest(pass, salt []byte) []byte {
	return pbkdf2.Key(pass, salt, Iter, sha256.Size, sha256.New)
}

func generatePassword() []byte {
	letterBytes := lib.Alphabets[Alphabet]

	b := make([]byte, KeyLen)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	return b
}
