package cmd

import "time"

const LimiterTime = 5 * time.Second

type SimpleRateLimiter struct {
	lastUserID string
	time       time.Time
}

func NewSimpleRateLimiter(lastUserID string, t time.Time) *SimpleRateLimiter {
	return &SimpleRateLimiter{lastUserID: lastUserID, time: t}
}

func (rl *SimpleRateLimiter) CheckLimit() (string, bool) {
	if time.Since(rl.time) <= LimiterTime {
		return rl.lastUserID, false
	}
	return " ", true
}

func (rl *SimpleRateLimiter) Unlock(UserID string) {
	rl.lastUserID = UserID
	rl.time = time.Now()
}
