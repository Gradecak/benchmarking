package ticker

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Ticker struct {
	*time.Ticker
}

func NewTicker(r time.Duration) *Ticker {
	return &Ticker{time.NewTicker(r)}
}

func (t *Ticker) SetRate(r time.Duration) {
	t.Ticker = time.NewTicker(r)
}

func (t Ticker) Tick(ctx context.Context, f func()) {
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
		go f()
	}
}

func (t Ticker) TickSync(ctx context.Context, f func(*sync.WaitGroup)) {
	wg := sync.WaitGroup{}
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			logrus.Info("Waiting for Lads to finish...")
			wg.Wait()
			return
		}
		wg.Add(1)
		go f(&wg)
	}
}
