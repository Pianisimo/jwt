package controllers

import (
	context2 "context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pianisimo/jwt/database"
	"github.com/pianisimo/jwt/helpers"
	"github.com/pianisimo/jwt/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	userCollection = database.OpenCollection(database.Client, "user")
	validate       = validator.New()
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func isPasswordValid(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(context *gin.Context) {
		timeoutCtx, cancelFunc := context2.WithTimeout(context2.Background(), 100*time.Second)
		var user models.User
		err := context.BindJSON(&user)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"BindJson": err.Error()})
			return
		}

		err = validate.Struct(user)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"validate struct": err.Error()})
			return
		}

		count, err := userCollection.CountDocuments(timeoutCtx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Error occurred while checking email"})
		}
		defer cancelFunc()

		password := HashPassword(*user.Password)
		user.Password = &password

		count2, err := userCollection.CountDocuments(timeoutCtx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Error occurred while phone number"})
		}

		if count > 0 || count2 > 0 {
			context.JSON(http.StatusBadRequest, gin.H{"error": "This email or phone number already exists"})
			return
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserId = user.ID.Hex()
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserId)
		user.Token = &token
		user.RefreshToken = &refreshToken

		result, err := userCollection.InsertOne(timeoutCtx, user)
		if err != nil {
			log.Panic(err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
		}

		defer cancelFunc()
		context.JSON(http.StatusOK, result)
	}
}

func Login() gin.HandlerFunc {
	return func(context *gin.Context) {
		timeoutCtx, cancelFunc := context2.WithTimeout(context2.Background(), 100*time.Second)

		var user models.User
		var foundUser models.User

		err := context.BindJSON(&user)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = userCollection.FindOne(timeoutCtx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancelFunc()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordIsValid, msg := isPasswordValid(*user.Password, *foundUser.Password)
		defer cancelFunc()

		if !passwordIsValid {
			context.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserId)

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "error creating token"})
			return
		}

		helpers.UpdateAllTokens(token, refreshToken, foundUser.UserId)
		err = userCollection.FindOne(timeoutCtx, bson.M{"user_id": foundUser.UserId}).Decode(&foundUser)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(context *gin.Context) {
		err := helpers.CheckUserType(context, "ADMIN")
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		timeoutCtx, cancelFunc := context2.WithTimeout(context2.Background(), 100*time.Second)
		defer cancelFunc()
		recordPerPage, err := strconv.Atoi(context.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(context.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(context.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := userCollection.Aggregate(timeoutCtx, mongo.Pipeline{
			matchStage, groupStage, projectStage})

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing user items"})
			return
		}
		defer cancelFunc()
		var allUsers []bson.M

		err = result.All(timeoutCtx, &allUsers)
		if err != nil {
			log.Fatal(err)
		}
		defer cancelFunc()
		context.JSON(http.StatusOK, allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		userId := context.Param("user_id")
		err := helpers.MatchUserTypeToUid(context, userId)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		timeoutCtx, cancelFunc := context2.WithTimeout(context2.Background(), 100*time.Second)
		defer cancelFunc()
		var user models.User

		err = userCollection.FindOne(timeoutCtx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.JSON(http.StatusOK, user)
	}
}
