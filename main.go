package main

import (
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/models"
	"itsky/a2b-api-go/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")
		if len(apiKey) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
			return
		}

		if apiKey != env.Env.ApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid API key"})
			return
		}

		c.Next()
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	api := r.Group("/api")
	api.Use(APIKeyAuthMiddleware())
	api.GET("/clientbalance", getClientBalance)
	api.POST("/clientrecharge", clientRecharge)

	return r
}

func main() {
	if err := models.ConnectDB(); err != nil {
		utils.Log.Fatalln(err)
	}
	defer models.DisconnectDB()
	r := setupRouter()
	r.Run(":" + env.Env.ApiPort)
}

// /api/clientbalance?kiraninumber=
func getClientBalance(c *gin.Context) {
	useralias := c.Query("kiraninumber")

	client := models.GetCard(useralias)
	if client == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, map[string]any{"credit": client.Credit})
}

// /api/clientrecharge?kiraninumber=&amount=&txRef=
func clientRecharge(c *gin.Context) {
	useralias := c.Query("kiraninumber")
	amount, err := strconv.Atoi(c.Query("amount"))
	txRef := c.Query("txRef")

	if len(useralias) != 11 || err != nil || amount < 0 || len(txRef) == 1 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client, err := models.CardRecharge(useralias, amount, txRef, time.Now())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if client == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, map[string]any{"credit": client.Credit})
}
