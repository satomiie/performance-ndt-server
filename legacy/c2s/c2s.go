package c2s

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/m-lab/ndt-server/legacy/metrics"
	"github.com/m-lab/ndt-server/legacy/protocol"
	"github.com/m-lab/ndt-server/legacy/testresponder"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	ready = float64(-1)
)

// Responder responds to c2s tests.
type Responder struct {
	testresponder.TestResponder
	Response chan float64
}

// TestServer performs the NDT c2s test.
func (tr *Responder) TestServer(w http.ResponseWriter, r *http.Request) {
	upgrader := testresponder.MakeNdtUpgrader([]string{"c2s"})
	wsc, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade should have already returned an HTTP error code.
		log.Println("ERROR C2S: upgrader", err)
		return
	}
	ws := protocol.AdaptWsConn(wsc)
	tr.performTest(ws)
}

func (tr *Responder) performTest(ws protocol.MeasuredConnection) {
	tr.Response <- ready
	bytesPerSecond := tr.recvC2SUntil(ws)
	tr.Response <- bytesPerSecond
	go func() {
		// After the test is supposedly over, let the socket drain a bit to not
		// confuse poorly-written clients by closing unexpectedly when there is still
		// buffered data. We make the judgement call that if the clients are so poorly
		// written that they still have data buffered after 5 seconds and are confused
		// when the c2s socket closes when buffered data is still in flight, then it
		// is okay to break them.
		ws.DrainUntil(time.Now().Add(5 * time.Second))
		ws.Close()
	}()
}

func (tr *Responder) recvC2SUntil(ws protocol.Connection) float64 {
	done := make(chan float64)

	go func() {
		startTime := time.Now()
		endTime := startTime.Add(10 * time.Second)
		totalBytes, err := ws.DrainUntil(endTime)
		if err != nil {
			tr.Close()
			return
		}
		bytesPerSecond := float64(totalBytes) / float64(time.Since(startTime)/time.Second)
		done <- bytesPerSecond
	}()

	log.Println("C2S: Waiting for test to complete or timeout")
	select {
	case <-tr.Ctx.Done():
		log.Println("C2S: Context Done!", tr.Ctx.Err())
		ws.Close()
		// Return zero on error.
		return 0
	case bytesPerSecond := <-done:
		return bytesPerSecond
	}
}

// ManageTest manages the c2s test lifecycle.
func ManageTest(ws protocol.Connection, config *testresponder.Config) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a testResponder instance.
	testResponder := &Responder{
		Response: make(chan float64),
	}
	testResponder.Config = config

	// Create a TLS server for running the C2S test.
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ndt_protocol",
		promhttp.InstrumentHandlerCounter(
			metrics.TestCount.MustCurryWith(prometheus.Labels{"direction": "c2s"}),
			http.HandlerFunc(testResponder.TestServer)))
	err := testResponder.StartAsync(ctx, serveMux, testResponder.performTest, "C2S")
	if err != nil {
		return 0, err
	}
	defer testResponder.Close()

	done := make(chan float64)
	go func() {
		// Wait for test to run.
		// Send the server port to the client.
		protocol.SendJSONMessage(protocol.TestPrepare, strconv.Itoa(testResponder.Port), ws)
		c2sReady := <-testResponder.Response
		if c2sReady != ready {
			log.Println("ERROR C2S: Bad value received on the c2s channel", c2sReady)
			cancel()
			return
		}
		// Tell the client to start the test.
		protocol.SendJSONMessage(protocol.TestStart, "", ws)

		// Wait for results to be generated.
		c2sBytesPerSecond := <-testResponder.Response
		c2sKbps := 8 * c2sBytesPerSecond / 1000.0

		// Finish the test.
		protocol.SendJSONMessage(protocol.TestMsg, fmt.Sprintf("%.4f", c2sKbps), ws)
		protocol.SendJSONMessage(protocol.TestFinalize, "", ws)
		log.Println("C2S: server rate:", c2sKbps)
		done <- c2sKbps
	}()

	select {
	case <-ctx.Done():
		log.Println("C2S: ctx Done!")
		return 0, ctx.Err()
	case value := <-done:
		log.Println("C2S: finished ", value)
		return value, nil
	}
}
