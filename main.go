package main

import (
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	api := r.Group("/api")
	api.GET("/clientbalance", getClientBalance)
	api.POST("/clientrecharge", clientRecharge)

	return r
}

func main() {
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

// /api/clientRecharge?kiraninumber=&amount=&txRef=
func clientRecharge(c *gin.Context) {
	useralias := c.Query("kiraninumber")
	amount, err := strconv.Atoi(c.Query("amount"))
	txRef := c.Query("txRef")

	if len(useralias) < 11 || err != nil || amount < 0 || len(txRef) == 1 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := models.CardRecharge(useralias, amount, txRef, time.Now())
	if client == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, map[string]any{"credit": client.Credit})
}
