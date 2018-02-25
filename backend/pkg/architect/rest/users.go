package rest

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
)

type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	Username string `json:"username"`
}

type userList struct {
	Total int             `json:"total"`
	Data  []*userResponse `json:"data"`
}

func newUserResponse(u *user.User) *userResponse {
	return &userResponse{
		Username: u.Username,
	}
}

type userHandler struct {
	userManager *user.Manager
}

func newUserHandler(userManager *user.Manager) *userHandler {
	return &userHandler{
		userManager: userManager,
	}
}

func (h *userHandler) create(c echo.Context) error {
	rU := new(userRequest)
	if err := c.Bind(rU); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		return nil
	}
	u, err := h.userManager.Create(rU.Username, rU.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.ErrorMap)
		return nil
	}

	c.JSON(http.StatusCreated, newUserResponse(u))
	return nil
}

func (h *userHandler) getAll(c echo.Context) error {
	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	us, total := h.userManager.GetAll(pQ)
	rUsers := []*userResponse{}
	for _, p := range us {
		rUsers = append(rUsers, newUserResponse(p))
	}

	c.JSON(http.StatusOK, userList{
		Total: total,
		Data:  rUsers,
	})
	return nil
}

func (h *userHandler) get(c echo.Context) error {
	if u := getUserByUsername(c, h.userManager); u != nil {
		c.JSON(http.StatusOK, newUserResponse(u))
	}

	return nil
}

// func (h *userHandler) update(c echo.Context) error {
// 	return nil
// }

func (h *userHandler) delete(c echo.Context) error {

	if u := getUserByUsername(c, h.userManager); u != nil {
		h.userManager.Delete(u)
		c.JSON(http.StatusOK, nil)
	}

	return nil
}

func getUserByUsername(c echo.Context, uM *user.Manager) *user.User {
	username := c.Param("username")

	u, err := uM.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return u
}
