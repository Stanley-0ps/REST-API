package routes

import (
	"errors"
	"net/http"

	"example.com/rest-api/models"
	"example.com/rest-api/utils"
	"github.com/gin-gonic/gin"
)

func signup(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	err = user.Save()
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken) {
			context.JSON(http.StatusConflict, gin.H{"message": "Email is already registered."})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user."})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully!"})
}

func login(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	err = user.ValidateCredentials()
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			context.JSON(http.StatusUnauthorized, gin.H{"message": "Could not authenticate user."})
			return
		}
		// Anything else is an unexpected failure (e.g. the database is
		// unreachable), so do not report it as a bad password.
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not process login."})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not authenticate user."})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}
