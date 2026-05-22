package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getUserIDFromContext(c *gin.Context) (primitive.ObjectID, bool) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		return primitive.NilObjectID, false
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		return primitive.NilObjectID, false
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return primitive.NilObjectID, false
	}

	return userID, true
}