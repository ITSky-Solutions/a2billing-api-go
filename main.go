package main

import (
	"fmt"
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/models"
	"itsky/a2b-api-go/utils"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

	exit := make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		if err := r.Run(":" + env.Env.ApiPort); err != nil {
			utils.Log.Println(err)
			exit <- syscall.Signal(0)
		}
	}()

	if !gin.IsDebugging() {
		utils.Log.Printf("Listening and serving HTTP on %s\n", env.Env.ApiPort)
	}
	<-exit
	utils.Log.Println("Shutting down server...")
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
	amount, err := strconv.ParseFloat(c.Query("amount"), 64)
	txRef := c.Query("txRef")

	if len(useralias) != 11 || err != nil || amount < 0 || len(txRef) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid parameters"})
		return
	}

	client, err := models.CardRecharge(useralias, amount, txRef, time.Now())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if client == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,
			gin.H{"message": fmt.Sprintf("Client %s not found", useralias)})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"credit": client.Credit})
}
