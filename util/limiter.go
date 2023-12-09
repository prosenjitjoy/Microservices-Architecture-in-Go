package util

import "golang.org/x/time/rate"

type limiter struct {
	l *rate.Limiter
}

func NewLimiter(limit, burst int) *limiter {
	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *limiter) Limit() bool {
	return l.l.Allow()
}
