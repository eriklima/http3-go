package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"

	"github.com/eriklima/http3-quic/utils"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

var certPath string

// var randomResponse0 *[]byte
var responses [4]*[]byte

func init() {
	setupCertPath()
}

func setupCertPath() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Join(path.Dir(filename), "keys")
}

func main() {
	addr := flag.String("addr", "localhost:4433", "Server listening to IP:PORT")
	qlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	qlogpath := flag.String("qlogpath", "qlog", "Custom path to save the qlog file. Require 'qlog'.")
	nbytes := flag.Int("bytes", 1_000_000, "Number of bytes to send to the server")
	flag.Parse()

	fmt.Println("Creating buffer...")

	// randomResponse = createBuf(1000 * 1000 * 1000)
	// randomResponse0 = createBuf(*nbytes)

	responses[0] = createBuf(*nbytes)
	responses[1] = createBuf(*nbytes * 2)
	responses[2] = createBuf(*nbytes * 4)
	responses[3] = createBuf(*nbytes * 8)

	// fmt.Printf("Buffer created with %d bytes\n", len(*randomResponse0))

	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(*addr, *qlog, *qlogpath, &wg)
	wg.Wait()

	fmt.Println("Server finished")
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

func startServer(addr string, enableQlog bool, qlogpath string, wg *sync.WaitGroup) {
	defer wg.Done()

	// TODO: testar com http.HandleFunc (igual ao HTTP2) ao invés de mux
	handler := setupHandler()
	quicConf := setupQuicConfig(enableQlog, qlogpath)

	server := http3.Server{
		Addr:       addr,
		Handler:    handler,
		QuicConfig: quicConf,
	}

	pem, key := getCertificatePaths()

	fmt.Printf("HTTP3 Server listening on %s\n", addr)

	err := server.ListenAndServeTLS(pem, key)

	if err != nil {
		fmt.Printf("Server error: %s\n", err)
	}
}

func setupHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request: ", r.Proto, r.Method, r.RequestURI)

		// reqBody, err := io.ReadAll(r.Body)

		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf("Request: %s Body: %d\n", r.RequestURI, len(reqBody))

		paramString := r.URL.Path[1:]
		paramNumber, err := strconv.ParseInt(paramString, 10, 0)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Param: %d\n", paramNumber)

		// w.Write([]byte(r.RequestURI))
		// w.Write([]byte(r.URL.String()))

		// w.Header().Add("X-Body-Size", fmt.Sprintf("%d", len(*randomResponse0)))
		// w.Write(*randomResponse0)

		response := *responses[paramNumber-1]

		w.Header().Add("X-Body-Size", fmt.Sprintf("%d", len(response)))
		w.Write(response)
	})

	return mux
}

func setupQuicConfig(enableQlog bool, qlogpath string) *quic.Config {
	config := &quic.Config{}

	if enableQlog {
		config.Tracer = func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			createQlogPath(qlogpath)
			filename := fmt.Sprintf("%s/server_%s.qlog", qlogpath, connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return qlog.NewConnectionTracer(utils.NewBufferedWriteCloser(bufio.NewWriter(f), f), p, connID)
		}

		fmt.Println("Qlog enabled!")
	}

	return config
}

func createQlogPath(qlogPath string) {
	err := os.MkdirAll(qlogPath, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func getCertificatePaths() (string, string) {
	// return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
	return "./keys/cert.pem", "./keys/priv.key"
}
