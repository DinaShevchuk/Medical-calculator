package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"dosage_drugs/internal/app/auth"
)

const (
	bearerPrefix = "Bearer "
	cookieName   = "access_token"

	ctxUserID   = "user_id"
	ctxToken    = "token"
	ctxTokenExp = "token_exp"
)

// extractToken берёт JWT из заголовка Authorization, а если его нет — из cookie.
func extractToken(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, bearerPrefix) {
		return strings.TrimSpace(h[len(bearerPrefix):])
	}
	if v, err := c.Cookie(cookieName); err == nil {
		return v
	}
	return ""
}

// authenticate возвращает HTTP-статус: 0 — токен валиден, 401 — гость/плохой токен,
// 500 — ошибка Redis.
func (h *Handler) authenticate(c *gin.Context) int {
	token := extractToken(c)
	if token == "" {
		return http.StatusUnauthorized
	}

	claims, err := auth.ParseToken(h.Config.JWTSecret, token)
	if err != nil {
		return http.StatusUnauthorized
	}

	blacklisted, err := h.Redis.IsBlacklisted(c.Request.Context(), token)
	if err != nil {
		logrus.Error("redis blacklist check: ", err)
		return http.StatusInternalServerError
	}
	if blacklisted {
		return http.StatusUnauthorized
	}

	c.Set(ctxUserID, claims.UserID)
	c.Set(ctxToken, token)
	c.Set(ctxTokenExp, claims.ExpiresAt.Time)
	return 0
}

// WithAuthCheck — обязательная авторизация (методы создателя).
func (h *Handler) WithAuthCheck(c *gin.Context) {
	switch h.authenticate(c) {
	case 0:
		c.Next()
	case http.StatusInternalServerError:
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"status": "error", "message": "ошибка проверки токена"})
	default:
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"status": "error", "message": "требуется авторизация"})
	}
}

// WithOptionalAuth — гость проходит дальше, но если токен валиден, пользователь
// попадает в контекст (нужно для is_mine / is_liked).
func (h *Handler) WithOptionalAuth(c *gin.Context) {
	if h.authenticate(c) == http.StatusInternalServerError {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"status": "error", "message": "ошибка проверки токена"})
		return
	}
	c.Next()
}

// currentUserID возвращает id авторизованного пользователя или 0 для гостя.
func currentUserID(c *gin.Context) int {
	return c.GetInt(ctxUserID)
}

func tokenAndTTL(c *gin.Context) (string, time.Duration) {
	token := c.GetString(ctxToken)
	exp, ok := c.Get(ctxTokenExp)
	if !ok {
		return token, 0
	}
	return token, time.Until(exp.(time.Time))
}
