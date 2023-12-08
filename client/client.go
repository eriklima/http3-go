package main

import (
	"crypto/rand"
	"flag"
	"log"
	"path"
	"runtime"
	"strconv"
	"sync"

	"github.com/eriklima/http3-quic/client/requesth3"
)

var certPath string
var loopCount int

func init() {
	setupCertPath()
}

func setupCertPath() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Dir(filename)
}

func main() {
	server := flag.String("server", "localhost:4433", "IP:PORT for HTTP3 server")
	qlog := flag.Bool("qlog", false, "Output a qlog (in the same directory)")
	qlogpath := flag.String("qlogpath", "qlog", "Custom path to save the qlog file. Require 'qlog'")
	nbytes := flag.Int("bytes", 0, "Number of bytes to send to the server")
	parallel := flag.Int("parallel", 1, "Number of parallel requests")
	experNumber := flag.Int("expernumber", 1, "Number of the experiment")
	flag.Parse()

	completedUrl := "https://" + *server

	buf := createBuf(*nbytes)

	loopCount = *parallel

	var wg sync.WaitGroup
	wg.Add(loopCount)
	for loopCount > 0 {
		// finalUrl := completedUrl + "/" + strconv.Itoa(loopCount)
		finalUrl := completedUrl + "/" + strconv.Itoa(*experNumber)

		go func(finalUrl string) {
			req := requesth3.RequestH3{
				finalUrl,
				certPath,
			}

			req.Execute(*buf, *qlog, *qlogpath)

			wg.Done()
		}(finalUrl)

		loopCount--
	}
	wg.Wait()

	// fmt.Println("Client finished.")
}

func createBuf(size int) *[]byte {
	buf := make([]byte, 0)

	if size > 0 {
		buf = make([]byte, size)

		// Randomize the buffer
		_, err := rand.Read(buf)

		if err != nil {
			log.Fatalf("error while generating random string: %s", err)
		}
	}

	return &buf
}
