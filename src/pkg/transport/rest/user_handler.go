package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type UserHandler interface {
	Signup(context *gin.Context)
	Login(context *gin.Context)
	GetUserById(context *gin.Context)
	GetAllUsers(context *gin.Context)
	UpdateUser(context *gin.Context)
	UpdatePassword(context *gin.Context)
	UpdateRoles(context *gin.Context)
	DeleteUser(context *gin.Context)
}

type userHandler struct {
	tokenVerifier TokenVerifier
	usecase       usecase.UserUsecase
}

func NewUserHandler(tokenVerifier TokenVerifier, usecase usecase.UserUsecase) UserHandler {
	return &userHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
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
	userId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only fetch his own data.
	// if he tries to get the data of another user he must be an admin.
	if authUserId != userId {
		hasAdminRole, err := token.HasRole(model.RoleAdmin)
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
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	hasAdminRole, err := token.HasRole(model.RoleAdmin)
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

type updateUserInput struct {
	Username string
}

func (handler *userHandler) UpdateUser(context *gin.Context) {
	userId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only update his own data.
	// if he tries to get the data of another user he must be an admin.
	if authUserId != userId {
		hasAdminRole, err := token.HasRole(model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !hasAdminRole {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to update this user"})
			return
		}
	}

	var userInput updateUserInput
	if err := context.ShouldBindJSON(&userInput); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingUser, err := handler.usecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user with id %v not found", userId)})
		return
	}
	existingUser.Username = userInput.Username

	err = handler.usecase.UpdateUser(existingUser)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v updated", userId)})
}

type passwordChangeInput struct {
	Password string
}

func (handler *userHandler) UpdatePassword(context *gin.Context) {
	userId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only update his own data.
	// if he tries to get the data of another user he must be an admin.
	if authUserId != userId {
		hasAdminRole, err := token.HasRole(model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !hasAdminRole {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to update this user"})
			return
		}
	}

	var passwordInput passwordChangeInput
	if err := context.ShouldBindJSON(&passwordInput); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(strings.TrimSpace(passwordInput.Password)) == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "password must not be empty"})
		return
	}

	err = handler.usecase.UpdateUserPassword(userId, passwordInput.Password)
	if err != nil {
		var entityNotFoundError *usecase.EntityNotFoundError
		var errorCode int
		switch {
		case errors.As(err, &entityNotFoundError):
			errorCode = http.StatusNotFound
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("password of user %v updated", userId)})
}

type rolesInput struct {
	Roles model.RoleList
}

func (handler *userHandler) UpdateRoles(context *gin.Context) {
	userId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	hasAdminRole, err := token.HasRole(model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !hasAdminRole {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to update this user"})
		return
	}
	var rolesInput rolesInput
	if err := context.ShouldBindJSON(&rolesInput); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := handler.usecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user with id %v does not exist", userId)})
		return
	}
	user.Roles = rolesInput.Roles
	err = handler.usecase.UpdateUser(user)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("roles of user %v updated", userId)})
}

func (handler *userHandler) DeleteUser(context *gin.Context) {
	userId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	hasAdminRole, err := token.HasRole(model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !hasAdminRole {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to delete a user"})
		return
	}
	_, err = handler.usecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user with id %v does not exist", userId)})
		return
	}
	err = handler.usecase.DeleteUser(userId)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v deleted", userId)})
}

func (handler *userHandler) getId(context *gin.Context) (uuid.UUID, error) {
	id := context.Param("id")
	if id == "" {
		return uuid.Nil, fmt.Errorf("please specify a valid id")
	}
	userId, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}
