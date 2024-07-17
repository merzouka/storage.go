package main

import (
	"log"
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
        filesArr := form.File
        names := []string{}
        for key, files := range filesArr {
            log.Println(key)
            names = append(names, key)
            for _, file := range files {
                ctx.SaveUploadedFile(file, "./hello")
            }
        }
        ctx.JSON(http.StatusOK, map[string][]string{
            "result": names,
        })
    })

    router.Run(":8080")
}
