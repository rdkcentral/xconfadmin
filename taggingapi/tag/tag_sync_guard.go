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
	interval := time.Second / time.Duration(perSecond)
	if interval < time.Nanosecond {
		interval = time.Nanosecond
	}
	return &rateLimiter{interval: interval}
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
// breaker instead. A missing member is a healthy XDAS answer, so it only
// breaks an error streak - it never trips the breaker on its own.
type syncBreaker struct {
	mu               sync.Mutex
	window           []syncOutcome
	idx              int
	filled           int
	total            int
	errInWindow      int
	consecErrors     int
	minSample        int
	errorRatePercent int
	maxConsecErrors  int
	trippedReason    string
}

// Fallbacks for a non-positive breaker knob, mirroring the config defaults.
// Every threshold is a >= comparison, so a zero is degenerate rather than
// lenient: it trips on the first healthy member, aborting the run blaming XDAS.
const (
	syncBreakerDefaultWindow          = 200
	syncBreakerDefaultErrorRate       = 25
	syncBreakerDefaultMaxConsecErrors = 10
)

func newSyncBreaker(window, minSample, errorRatePercent, maxConsecErrors int) *syncBreaker {
	if window <= 0 {
		window = syncBreakerDefaultWindow
	}
	if errorRatePercent <= 0 {
		errorRatePercent = syncBreakerDefaultErrorRate
	}
	if maxConsecErrors <= 0 {
		maxConsecErrors = syncBreakerDefaultMaxConsecErrors
	}
	// Zero minSample is coherent (arm from the first member); only fix
	// negatives.
	if minSample < 0 {
		minSample = 0
	}
	return &syncBreaker{
		window:           make([]syncOutcome, window),
		minSample:        minSample,
		errorRatePercent: errorRatePercent,
		maxConsecErrors:  maxConsecErrors,
	}
}

func (b *syncBreaker) record(o syncOutcome) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.filled == len(b.window) {
		if b.window[b.idx] == syncOutcomeError {
			b.errInWindow--
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
	// The error rate arms only after minSample members total; the rate
	// itself is computed over the sliding window (filled entries).
	if b.total >= b.minSample && b.errInWindow*100 >= b.errorRatePercent*b.filled {
		b.trippedReason = "xdas_unhealthy_error_rate"
	}
}

func (b *syncBreaker) tripped() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.trippedReason, b.trippedReason != ""
}
