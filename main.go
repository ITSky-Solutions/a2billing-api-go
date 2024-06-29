package main

import (
	"net/http"
	"strconv"

	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/models"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/clientbalance", getClientBalance)
	r.POST("/clientrecharge", clientRecharge)

	return r
}

func main() {
	r := setupRouter()
	r.Run(":"+env.Env.ApiPort)
}

func getClientBalance(c *gin.Context)  {
	useralias := c.Query("kiraninumber");

	client := models.GetCard(useralias)
	if client == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, client)
}

func clientRecharge(c *gin.Context)  {
	useralias := c.Query("kiraninumber");
	amount, err := strconv.Atoi(c.Query("amount"));

	if len(useralias) < 11 || err != nil || amount < 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := models.CardRecharge(useralias, amount)
	if client == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, client)
}
