package tag

import (
	"context"
	"sync"
	"time"
)

// rateLimiter spaces calls evenly so the whole run never exceeds the
// configured XDAS calls/sec, no matter how many workers share it.
type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func newRateLimiter(perSecond int) *rateLimiter {
	if perSecond <= 0 {
		perSecond = 1
	}
	return &rateLimiter{interval: time.Second / time.Duration(perSecond)}
}

func (l *rateLimiter) wait(ctx context.Context) error {
	l.mu.Lock()
	now := time.Now()
	if l.next.Before(now) {
		l.next = now
	}
	sleep := l.next.Sub(now)
	l.next = l.next.Add(l.interval)
	l.mu.Unlock()

	if sleep <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(sleep)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type syncOutcome int

const (
	syncOutcomeOk syncOutcome = iota
	syncOutcomeMissing
	syncOutcomeError
)

// syncBreaker watches a sliding window of per-member outcomes and stops the
// run when XDAS looks unhealthy. A transport error or 5xx must never be
// counted as missing; enough of them in a row (or in the window) trips the
// breaker instead. A suspiciously high missing rate does not trip by itself
// - it raises missingHigh so the engine can confirm the members are genuinely
// missing with a probe GET on a known-good member before continuing.
type syncBreaker struct {
	mu                 sync.Mutex
	window             []syncOutcome
	idx                int
	filled             int
	total              int
	errInWindow        int
	missInWindow       int
	consecErrors       int
	minSample          int
	errorRatePercent   int
	missingRatePercent int
	maxConsecErrors    int
	trippedReason      string
	missingHigh        bool
}

func newSyncBreaker(window, minSample, errorRatePercent, missingRatePercent, maxConsecErrors int) *syncBreaker {
	if window <= 0 {
		window = 200
	}
	return &syncBreaker{
		window:             make([]syncOutcome, window),
		minSample:          minSample,
		errorRatePercent:   errorRatePercent,
		missingRatePercent: missingRatePercent,
		maxConsecErrors:    maxConsecErrors,
	}
}

func (b *syncBreaker) record(o syncOutcome) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.filled == len(b.window) {
		switch b.window[b.idx] {
		case syncOutcomeError:
			b.errInWindow--
		case syncOutcomeMissing:
			b.missInWindow--
		}
	} else {
		b.filled++
	}
	b.window[b.idx] = o
	b.idx = (b.idx + 1) % len(b.window)
	b.total++

	switch o {
	case syncOutcomeError:
		b.errInWindow++
		b.consecErrors++
	case syncOutcomeMissing:
		b.missInWindow++
		b.consecErrors = 0
	default:
		b.consecErrors = 0
	}

	if b.trippedReason != "" {
		return
	}
	if b.consecErrors >= b.maxConsecErrors {
		b.trippedReason = "xdas_unhealthy_consecutive_errors"
		return
	}
	// Rate thresholds arm only after minSample members total; the rates
	// themselves are computed over the sliding window (filled entries).
	if b.total >= b.minSample {
		if b.errInWindow*100 >= b.errorRatePercent*b.filled {
			b.trippedReason = "xdas_unhealthy_error_rate"
			return
		}
		if b.missInWindow*100 >= b.missingRatePercent*b.filled {
			b.missingHigh = true
		}
	}
}

func (b *syncBreaker) tripped() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.trippedReason, b.trippedReason != ""
}

func (b *syncBreaker) missingRateHigh() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.missingHigh
}

// clearMissingHigh is called after a probe confirmed the missing members are
// real, so the high missing rate stops being treated as a suspected outage.
func (b *syncBreaker) clearMissingHigh() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.missingHigh = false
}
