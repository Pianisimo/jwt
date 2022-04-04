package helpers

import (
	"errors"
	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUid(context *gin.Context, userId string) (err error) {
	userType := context.GetString("user_type")
	uid := context.GetString("uid")
	err = nil

	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	err = CheckUserType(context, userType)
	return err
}

func CheckUserType(context *gin.Context, role string) (err error) {
	userType := context.GetString("user_type")
	err = nil

	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	return err
}
