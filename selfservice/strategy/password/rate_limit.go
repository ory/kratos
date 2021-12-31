package password

import (
	"crypto/sha1"
	"fmt"
	"github.com/ory/kratos/driver/config"
	"golang.org/x/time/rate"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

var visitors = make(map[string]*visitor)
var mu sync.Mutex

type visitor struct {
	limiter  *rate.Limiter
	backoff  time.Duration
	lastSeen time.Time
}

func getRequestIp(r *http.Request) string {
	ipAddress := r.RemoteAddr
	fwdAddress := r.Header.Get("X-Forwarded-For")
	if fwdAddress != "" {
		ipAddress = fwdAddress
		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ipAddress = ips[0]
		}
	}
	return ipAddress
}

func getVisitorKey(r *http.Request, p submitSelfServiceLoginFlowWithPasswordMethodBody) string {
	keyString := fmt.Sprintf("%s-%s", p.Identifier, getRequestIp(r))
	hash := sha1.Sum([]byte(keyString))
	return fmt.Sprintf("%x", hash)
}

func getVisitor(r *http.Request, p submitSelfServiceLoginFlowWithPasswordMethodBody) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	visitorKey := getVisitorKey(r, p)
	v, exists := visitors[visitorKey]
	// reset the clock after some reasonable amount of time, if it hasn't been cleaned up otherwise
	if !exists || time.Now().Sub(v.lastSeen) > time.Minute*10 {
		backoff, _ := time.ParseDuration("1s")
		l := rate.NewLimiter(rate.Every(backoff), 2)
		visitors[visitorKey] = &visitor{l, backoff, time.Now()}
		return l
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func increaseRateLimitWait(config *config.Config, r *http.Request, p submitSelfServiceLoginFlowWithPasswordMethodBody) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	passwordPolicyConfig := config.PasswordPolicyConfig()

	visitorKey := getVisitorKey(r, p)
	v, exists := visitors[visitorKey]
	if exists {
		max, _ := time.ParseDuration(passwordPolicyConfig.RateLimitMaxDuration)
		v.backoff = time.Second * time.Duration(math.Min(v.backoff.Seconds()*2.0, max.Seconds()))
		v.limiter.SetLimit(rate.Every(v.backoff))
		v.lastSeen = time.Now()
		return v.limiter
	}
	return nil
}

func CleanupRateLimits() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
