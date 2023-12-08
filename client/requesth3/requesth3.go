package requesth3

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"

	"github.com/eriklima/http3-quic/utils"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

type RequestH3 struct {
	FinalUrl string
	// RoundTripper http3.RoundTripper
	// Pool x509.CertPool
	CertPath string
}

type Metrics struct {
	dnsLookup        time.Duration
	tcpConnection    time.Duration
	tlsHandshake     time.Duration
	serverProcessing time.Duration
	contentTransfer  time.Duration
	connect          time.Duration
	preTransfer      time.Duration
	startTransfer    time.Duration
	total            time.Duration
}

func (r *RequestH3) Execute(
	buf []byte,
	qlog bool,
	qlogpath string) {

	// var response *http.Response
	var req *http.Request
	var err error

	// fmt.Printf("Call %s\n", r.FinalUrl)

	if len(buf) == 0 {
		// response, err = client.Get(r.FinalUrl)
		req, err = http.NewRequest("GET", r.FinalUrl, nil)
		if err != nil {
			log.Fatal("GET Request error: ", err)
		}
	} else {
		// response, err = client.Post(r.FinalUrl, "application/octet-stream", bytes.NewReader(buf))
		req, err = http.NewRequest("POST", r.FinalUrl, bytes.NewReader(buf))
		if err != nil {
			log.Fatal("POST Request error: ", err)
		}
	}

	// fmt.Printf("METHOD :::: %s\n", req.Method)

	// ------------------------------------------------------------

	tInit := time.Now()

	var tDNSStart, tConnectStart, tConnectDone, tGotConn, tGotFirstResponseByte, tTLSHandshakeStart, tTLSHandshakeDone, tGetConn time.Time

	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) {
			if tConnectStart.IsZero() {
				// connecting to IP
				tConnectStart = time.Now()
			}
			// fmt.Println("EVENT: ConnectStart", tConnectStart.Sub(tInit))
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			tConnectDone = time.Now()

			// fmt.Println("EVENT: ConnectDone", tConnectDone.Sub(tInit))
		},
		DNSStart: func(_ httptrace.DNSStartInfo) { tDNSStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { tConnectStart = time.Now() },
		GetConn: func(_ string) {
			tGetConn = time.Now()
			// fmt.Println("EVENT: GetConn", tGetConn.Sub(tInit))
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			tGotConn = time.Now()
			// fmt.Println("EVENT: GotConn", tGotConn.Sub(tInit))
		},
		GotFirstResponseByte: func() {
			tGotFirstResponseByte = time.Now()
			// fmt.Println("EVENT: GotFirstResponseByte", tGotFirstResponseByte.Sub(tInit))
		},
		TLSHandshakeStart: func() {
			tTLSHandshakeStart = time.Now()
			fmt.Println("EVENT: TLSHandshakeStart", tTLSHandshakeStart.Sub(tInit))
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			tTLSHandshakeDone = time.Now()
			if err != nil {
				log.Fatalf("failed to perform TLS handshake: %v", err)
			}
			fmt.Println("EVENT: TLSHandshakeDone", tTLSHandshakeDone.Sub(tInit))
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// ------------------------------------------------------------

	pool := getCertPool()
	addRootCA(pool, r.CertPath)

	var keyLog io.Writer

	tlsConfig := &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: true,
		KeyLogWriter:       keyLog,
	}

	quicConfig := setupQuicConfig(qlog, qlogpath)

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig:      quicConfig,
	}

	// TODO: colocar o defer para fora da função
	defer roundTripper.Close()

	client := &http.Client{
		Transport: roundTripper,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Request error: ", err)
	}

	// // TODO: remover depois dos testes
	// print("Sleeping...")
	// time.Sleep(time.Second * 5)
	// print("Done\n")

	body := getBody(res)

	tAfterReadBody := time.Now() // after read body
	// fmt.Println("EVENT: ResponseReady", tAfterReadBody.Sub(tInit))

	// fmt.Printf("Resposta: %#v\n", res)

	// responseLength := res.ContentLength

	bodySize := res.Header.Get("X-Body-Size")

	res.Body.Close()

	// fmt.Printf("Body: %s\n", body)

	// tAfterReadBody := time.Now() // after read body

	dnsLookup := tConnectStart.Sub(tDNSStart) // dns lookup
	// tcpConnection := tConnectDone.Sub(tConnectStart)             // tcp connection
	tcpConnection := tGotConn.Sub(tGetConn)                   // tcp connection
	tlsHandshake := tTLSHandshakeDone.Sub(tTLSHandshakeStart) // tls handshake
	serverProcessing := tGotFirstResponseByte.Sub(tGotConn)   // server processing
	contentTransfer := tAfterReadBody.Sub(tGotConn)           // content transfer
	connect := tConnectDone.Sub(tDNSStart)                    // connect
	preTransfer := tGotConn.Sub(tDNSStart)                    // pretransfer
	startTransfer := tGotFirstResponseByte.Sub(tDNSStart)     // starttransfer
	// total := t7.Sub(t0)            // total
	// tTotal := t7.Sub(t3) // total
	// tTotal := t7.Sub(tGetConn) // total
	// tTotal := tGotConn.Sub(tGetConn) // total
	tTotal := tAfterReadBody.Sub(tGetConn) // total
	// tTotal := tFinish.Sub(tStart)

	metrics := &Metrics{
		dnsLookup:        dnsLookup,
		tcpConnection:    tcpConnection,
		tlsHandshake:     tlsHandshake,
		serverProcessing: serverProcessing,
		contentTransfer:  contentTransfer,
		connect:          connect,
		preTransfer:      preTransfer,
		startTransfer:    startTransfer,
		// total:            total,
		total: tTotal,
	}

	saveMetrics(metrics, r.CertPath)

	fmt.Printf("\nProtocol: %s %s\n", res.Proto, req.Method)
	fmt.Printf("Code: %d\n", res.StatusCode)
	fmt.Printf("Content-Length: %d\n", res.ContentLength)
	fmt.Printf("Body: %d\n", len(body))
	// fmt.Printf("Body Length: %d\n\n", responseLength)
	fmt.Printf("Body Size Header: %s\n", bodySize)

	// fmt.Printf("\nDNS Lookup: %s\n", dnsLookup)
	fmt.Printf("\nConnection time: %s\n", tcpConnection)
	// fmt.Printf("TLS handshake: %s\n", tlsHandshake)
	// fmt.Printf("Server processing: %s\n", serverProcessing)
	fmt.Printf("Content transfer: %s\n", contentTransfer)
	// fmt.Printf("Connection: %s\n", connect)
	// fmt.Printf("Pretransfer: %s\n", preTransfer)
	// fmt.Printf("Starttransfer: %s\n", startTransfer)
	// fmt.Printf("Total: %s\n", t7.Sub(t0))
	fmt.Printf("Total: %s\n", tTotal)
	fmt.Print("-----------------------------------------\n")
}

func getCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()

	if err != nil {
		log.Fatal(err)
	}

	return pool
}

func addRootCA(certPool *x509.CertPool, certPath string) {
	// caCertPath := path.Join(certPath, "ca.pem")
	caCertPath := "./ca.pem"
	caCertRaw, err := os.ReadFile(caCertPath)
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		panic("FAILURE: Could not add root ceritificate to pool.")
	}
}

func setupQuicConfig(enableQlog bool, qlogPath string) *quic.Config {
	config := &quic.Config{}

	if enableQlog {
		config.Tracer = func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			createQlogPath(qlogPath)
			filename := fmt.Sprintf("%s/client_%s.qlog", qlogPath, connID)
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

func getBody(response *http.Response) []byte {
	body := &bytes.Buffer{}

	_, err := io.Copy(body, response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return body.Bytes()
}

func saveMetrics(metrics *Metrics, csvPath string) {
	// f, err := os.OpenFile(path.Join(csvPath, "metrics.csv"), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	f, err := os.OpenFile("/logs/metrics.csv", os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	if err != nil {
		// fmt.Println(err)
		// return
		log.Fatal(err)
	}
	w := csv.NewWriter(f)
	// for i := 0; i < 10; i++ {
	// 	w.Write([]string{"a", "b", "c"})
	// }
	row := []string{
		metrics.dnsLookup.String(),
		metrics.tcpConnection.String(),
		metrics.tlsHandshake.String(),
		metrics.serverProcessing.String(),
		metrics.contentTransfer.String(),
		metrics.connect.String(),
		metrics.preTransfer.String(),
		metrics.startTransfer.String(),
		metrics.total.String(),
	}

	w.Write(row)

	w.Flush()
}
