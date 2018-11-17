package rest

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Username string    `json:"username"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires"`
}

func newAuthResponse(u *user.User) *authResponse {
	sessionDuration := time.Hour * 24 * 2
	token, expires := auth.NewJWT(sessionDuration, auth.AudienceUser, u.ID)

	return &authResponse{
		Username: u.Username,
		Token:    token,
		Expires:  expires,
	}
}

type authHandler struct {
	userManager *user.Manager
}

func newAuthHandler(userManager *user.Manager) *authHandler {
	return &authHandler{
		userManager: userManager,
	}
}

func (h *authHandler) create(c echo.Context) error {
	rU := new(authRequest)
	if err := c.Bind(rU); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}

	u, err := h.userManager.GetByUsername(rU.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
		return nil
	}
	if !u.ValidatePassword(rU.Password) {
		c.JSON(http.StatusUnauthorized, nil)
		return nil
	}
	c.JSON(http.StatusCreated, newAuthResponse(u))
	return nil
}
