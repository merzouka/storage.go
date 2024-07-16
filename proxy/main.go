package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)



func main() {
    router := gin.Default()

    router.POST("/upload", func(ctx *gin.Context) {
        form, err := ctx.MultipartForm()
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "message": "please provide valid data",
            })
        }
        files := form.File
    })

    router.Run(":8080")
}
