package http

import (
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	Signup(context *gin.Context)
	Login(context *gin.Context)
}

type userHandler struct {
	usecase usecase.UserUsecase
}

func NewUserHandler(usecase usecase.UserUsecase) UserHandler {
	return &userHandler{
		usecase: usecase,
	}
}

type signupInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (handler *userHandler) Signup(context *gin.Context) {
	var input signupInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := model.User{
		Username: input.Username,
		Password: input.Password,
	}
	createdUser, err := handler.usecase.AddUser(&user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	createdUser.Password = ""
	context.JSON(http.StatusOK, createdUser)
}

type loginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (handler *userHandler) Login(context *gin.Context) {
	var input loginInput

	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.checkLogin(input.Username, input.Password)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "username or password is incorrect."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"token": token})
}

func (handler *userHandler) checkLogin(username string, password string) (string, error) {
	user, err := handler.usecase.GetUserByName(username)
	if err != nil {
		return "", err
	}
	err = handler.usecase.VerifyPassword(password, user.Password)
	if err != nil {
		return "", err
	}
	token, err := GenerateToken(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
