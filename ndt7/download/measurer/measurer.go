// Package measurer contains the downloader measurer
package measurer

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt-server/bbr"
	"github.com/m-lab/ndt-server/fdcache"
	"github.com/m-lab/ndt-server/logging"
	"github.com/m-lab/ndt-server/ndt7/model"
	"github.com/m-lab/ndt-server/ndt7/spec"
	"github.com/m-lab/ndt-server/tcpinfox"
)

func getSocketAndPossiblyEnableBBR(conn *websocket.Conn) (*os.File, error) {
	fp := fdcache.GetAndForgetFile(conn.UnderlyingConn())
	// Implementation note: in theory fp SHOULD always be non-nil because
	// now we always register the fp bound to a net.TCPConn. However, in
	// some weird cases it MAY happen that the cache pruning mechanism will
	// remove the fp BEFORE we can steal it. In case we cannot get a file
	// we just abort the test, as this should not happen (TM).
	if fp == nil {
		return nil, errors.New("cannot get file bound to websocket conn")
	}
	err := bbr.Enable(fp)
	if err != nil {
		logging.Logger.WithError(err).Warn("Cannot enable BBR")
		// FALLTHROUGH
	}
	return fp, nil
}

func measure(measurement *model.Measurement, sockfp *os.File) {
	bbrinfo, err := bbr.GetMaxBandwidthAndMinRTT(sockfp)
	if err == nil {
		measurement.BBRInfo = &bbrinfo
	}
	metrics, err := tcpinfox.GetTCPInfo(sockfp)
	if err == nil {
		measurement.TCPInfo = &metrics
	}
}

func loop(ctx context.Context, conn *websocket.Conn, dst chan<- model.Measurement) {
	logging.Logger.Debug("measurer: start")
	defer logging.Logger.Debug("measurer: stop")
	defer close(dst)
	measurerctx, cancel := context.WithTimeout(ctx, spec.DefaultRuntime)
	defer cancel()
	sockfp, err := getSocketAndPossiblyEnableBBR(conn)
	if err != nil {
		logging.Logger.WithError(err).Warn("getSocketAndPossiblyEnableBBR failed")
		return
	}
	defer sockfp.Close()
	t0 := time.Now()
	ticker := time.NewTicker(spec.MinMeasurementInterval)
	defer ticker.Stop()
	for {
		select {
		case <-measurerctx.Done(): // Liveness!
			logging.Logger.Debug("measurer: context done")
			return
		case now := <-ticker.C:
			elapsed := now.Sub(t0)
			measurement := model.Measurement{
				Elapsed: elapsed.Seconds(),
			}
			measure(&measurement, sockfp)
			dst <- measurement // Liveness: this is blocking
		}
	}
}

// Start runs the measurement loop in a background goroutine and emits
// the measurements on the returned channel.
//
// Liveness guarantee: the measurer will always terminate after
// a timeout of DefaultRuntime seconds, provided that the consumer
// continues reading from the returned channel.
func Start(ctx context.Context, conn *websocket.Conn) <-chan model.Measurement {
	dst := make(chan model.Measurement)
	go loop(ctx, conn, dst)
	return dst
}
