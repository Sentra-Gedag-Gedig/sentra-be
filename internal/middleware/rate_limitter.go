package middleware

import (
	"ProjectGolang/pkg/response"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
)

var (
	ErrTooManyRequests = response.NewError(http.StatusTooManyRequests, "too many requests")
)

type rateLimiter struct {
	bucket    map[string]*rate.Limiter
	rate      rate.Limit
	burstSize int
	mutex     *sync.RWMutex
}

func newRateLimiter(reqRate rate.Limit, burstSize int) *rateLimiter {
	return &rateLimiter{
		bucket:    make(map[string]*rate.Limiter),
		rate:      reqRate,
		burstSize: burstSize,
		mutex:     &sync.RWMutex{},
	}
}

func (r *rateLimiter) GetLimiterFrom(ip string) *rate.Limiter {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exist := r.bucket[ip]; !exist {
		r.bucket[ip] = rate.NewLimiter(r.rate, r.burstSize)
	}

	return r.bucket[ip]
}

func (m *middleware) NewRateLimiter(ctx *fiber.Ctx) error {
	clientIP := ctx.IP()
	limiter := m.rateLimitter.GetLimiterFrom(clientIP)

	if !limiter.Allow() {
		m.log.Warnf("too many requests for IP %s", clientIP)
		return ctx.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Too many requests",
		})
	}

	return ctx.Next()
}
