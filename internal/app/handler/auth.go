package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"dosage_drugs/internal/app/auth"
)

// ---------- DTO (используются и в swagger) ----------

type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"описание ошибки"`
}

type SuccessResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message,omitempty" example:"выполнено"`
}

type RegisterRequest struct {
	Login    string `json:"login" example:"ivanov"`
	Password string `json:"password" example:"secret123"`
}

type LoginRequest struct {
	Login    string `json:"login" example:"test"`
	Password string `json:"password" example:"test123"`
}

type UserInfo struct {
	ID    int    `json:"id" example:"2"`
	Login string `json:"login" example:"ivanov"`
}

type RegisterResponse struct {
	Status string   `json:"status" example:"success"`
	Data   UserInfo `json:"data"`
}

type LoginData struct {
	AccessToken string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string   `json:"token_type" example:"Bearer"`
	ExpiresIn   int      `json:"expires_in" example:"3600"`
	User        UserInfo `json:"user"`
}

type LoginResponse struct {
	Status string    `json:"status" example:"success"`
	Data   LoginData `json:"data"`
}

// ---------- Регистрация ----------

// APIRegister godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя. Пароль хранится в виде bcrypt-хеша. Доступно гостю.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Логин и пароль"
// @Success      201   {object}  RegisterResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /users/register [post]
func (h *Handler) APIRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "некорректный JSON"})
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if len(req.Login) < 3 || len(req.Login) > 25 {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "login: от 3 до 25 символов"})
		return
	}
	if len(req.Password) < 6 || len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "password: от 6 до 72 символов"})
		return
	}

	if _, err := h.Repository.GetUserByLogin(req.Login); err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{"error", "пользователь с таким логином уже существует"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ошибка базы данных"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ошибка хеширования пароля"})
		return
	}

	user, err := h.Repository.RegisterUser(req.Login, string(hash))
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "не удалось создать пользователя"})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		Status: "success",
		Data:   UserInfo{ID: user.ID, Login: user.Login},
	})
}

// ---------- Аутентификация ----------

// APILogin godoc
// @Summary      Аутентификация
// @Description  Проверяет логин и пароль, выдаёт JWT. Токен возвращается в теле ответа и дополнительно ставится в HttpOnly-cookie access_token. Доступно гостю.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Логин и пароль"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /users/login [post]
func (h *Handler) APILogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{"error", "login и password обязательны"})
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "неверный логин или пароль"})
			return
		}
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ошибка базы данных"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "неверный логин или пароль"})
		return
	}

	token, _, err := auth.GenerateToken(h.Config.JWTSecret, h.Config.JWTExpiresIn, user.ID, user.Login)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "не удалось создать токен"})
		return
	}

	maxAge := int(h.Config.JWTExpiresIn.Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieName, token, maxAge, "/", "", false, true)

	c.JSON(http.StatusOK, LoginResponse{
		Status: "success",
		Data: LoginData{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   maxAge,
			User:        UserInfo{ID: user.ID, Login: user.Login},
		},
	})
}

// ---------- Деавторизация ----------

// APILogout godoc
// @Summary      Деавторизация
// @Description  Помещает текущий JWT в blacklist Redis (до истечения его срока) и сбрасывает cookie. Доступно создателю.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/logout [post]
func (h *Handler) APILogout(c *gin.Context) {
	token, ttl := tokenAndTTL(c)

	if err := h.Redis.BlacklistToken(c.Request.Context(), token, ttl); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "не удалось выполнить выход"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieName, "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, SuccessResponse{Status: "success", Message: "выход выполнен"})
}
