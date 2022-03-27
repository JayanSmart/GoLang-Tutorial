package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jayansmart/GoLang-Tutorial/go-rest-api/controllers"
	"github.com/jayansmart/GoLang-Tutorial/go-rest-api/models"
)

func main() {
	r := gin.Default()

	models.ConnectDatabase()

	r.GET("/books", controllers.FindBooks)
	r.GET("/books/:id", controllers.FindBook)
	r.POST("/books", controllers.CreateBook)
	r.PATCH("/books/:id", controllers.UpdateBook)
	r.DELETE("/books/:id", controllers.DeleteBook)

	if err := r.Run(); err != nil {
		panic("Application was unable to start: " + err.Error())
	}
}
