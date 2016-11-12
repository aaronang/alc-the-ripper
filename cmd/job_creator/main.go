package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"net"
	"net/http"
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
	Iter      = 2000
)

func main() {
	ip := pflag.IP("ip", net.IPv4(127, 0, 0, 1), "Master IP")
	port := pflag.String("port", "8080", "Web server port")
	jobs := pflag.Int("jobs", 20, "Number of jobs to create")
	interval := pflag.Int("interval", 60, "Interval between job creation in seconds")
	pflag.Parse()

	addr := lib.Protocol + net.JoinHostPort(ip.String(), *port) + lib.JobsCreatePath

	for i := 0; i < *jobs; i++ {
		if resp, err := createJob(addr); err != nil || resp.StatusCode != http.StatusOK {
			log.Panicln("[createJob] Job wasn't created properly.", err, resp.Status)
		}
		log.Println("[createJob] Job was created successfully.")
		time.Sleep(time.Duration(*interval) * time.Second)
	}

}

func createJob(addr string) (*http.Response, error) {
	body := generateJobJSON()
	return http.Post(addr, lib.BodyType, bytes.NewBuffer(body))
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
