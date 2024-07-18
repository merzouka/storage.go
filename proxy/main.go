package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)



func main() {
    router := gin.Default()
    router.Use(cors.Default())

    router.POST("/upload", func(ctx *gin.Context) {
        form, err := ctx.MultipartForm()
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "message": "please provide valid data",
            })
        }
        filesArr := form.File
        names := []string{}
        for key, files := range filesArr {
            names = append(names, key)
            for _, file := range files {
                ctx.SaveUploadedFile(file, fmt.Sprintf("./files/%s", key))
            }
        }
        ctx.JSON(http.StatusOK, map[string][]string{
            "uploaded": names,
        })
    })

    router.Run(":8080")
}
