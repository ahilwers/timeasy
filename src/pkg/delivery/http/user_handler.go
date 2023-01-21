package http

import (
	"errors"
	"fmt"
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type UserHandler interface {
	Signup(context *gin.Context)
	Login(context *gin.Context)
	GetUserById(context *gin.Context)
	GetAllUsers(contest *gin.Context)
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
		var entityExistsError *usecase.EntityExistsError
		var errorCode int
		switch {
		case errors.As(err, &entityExistsError):
			errorCode = http.StatusConflict
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
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
	token, err := GenerateToken(user.ID, user.Roles)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (handler *userHandler) GetUserById(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "please specify a valid id"})
		return
	}
	userId, err := uuid.FromString(id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := ExtractToken(context)
	authUserId, err := ExtractTokenUserId(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only fetch his own data.
	// if he tries to get the data of another user he must be an admin.
	if authUserId != userId {
		hasAdminRole, err := TokenHasRole(token, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !hasAdminRole {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to fetch this user"})
			return
		}
	}
	user, err := handler.usecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user with id %v not found", userId)})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (handler *userHandler) GetAllUsers(context *gin.Context) {
	token := ExtractToken(context)
	hasAdminRole, err := TokenHasRole(token, model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !hasAdminRole {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to fetch this user"})
		return
	}

	users, err := handler.usecase.GetAllUsers()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all users"})
		return
	}
	context.JSON(http.StatusOK, users)
}
