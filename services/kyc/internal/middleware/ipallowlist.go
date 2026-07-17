package middleware

import (
	"net"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type IPAllowlist struct {
	allowed []net.IPNet
	logger  *zap.Logger
}

func NewIPAllowlist(cidrs []string, logger *zap.Logger) *IPAllowlist {
	nets := make([]net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			nets = append(nets, *ipNet)
		} else {
			logger.Warn("invalid CIDR in IP allowlist", zap.String("cidr", cidr))
		}
	}
	return &IPAllowlist{allowed: nets, logger: logger}
}

func (a *IPAllowlist) Middleware(c fiber.Ctx) error {
	if len(a.allowed) == 0 {
		return c.Next()
	}

	clientIP := net.ParseIP(c.IP())
	if clientIP == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success":   false,
			"error":     "KYC_SEC_001: IP invalide",
			"timestamp": c.Context().Value("timestamp"),
		})
	}

	for _, allowed := range a.allowed {
		if allowed.Contains(clientIP) {
			return c.Next()
		}
	}

	a.logger.Warn("webhook blocked — IP not in allowlist",
		zap.String("ip", c.IP()),
		zap.String("path", c.Path()))

	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"success": false,
		"error":   "KYC_SEC_002: IP source non autorisée",
	})
}
