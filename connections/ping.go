package connections

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	select {
	case <-ctx.Done():
		return
	case interval = <-reset: //Pulled initial interval off reset channel
	default:
	}

	if interval <= 0 {
		interval = defaultPingInterval
	}

	timer := time.NewTimer(interval)

	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		case <-ctx.Done(): // Context is canceled
			return
		case newInterval := <-reset: // Receive new duration from rest channel
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C: // Timer expires
			if _, err := w.Write([]byte("ping")); err != nil {
				// Track and act on consecutive timeouts
				return
			}
		}
		timer.Reset(interval)

	}
}
