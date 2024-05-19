package api

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Upload(file *multipart.FileHeader) (string, error) {
	var settings = GetSettings()

	src, err := file.Open()
	if err != nil {
		return "", err
	}

	defer src.Close()
	n := file.Filename
	dst := fmt.Sprintf("%s/%s", settings.StorePath, n)
	out, err := os.Create(dst)

	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, src)

	return n, err
}
func Download(n string) (string, []byte, error) {
	var settings = GetSettings()

	dst := fmt.Sprintf("%s/%s", settings.StorePath, n)
	b, err := os.ReadFile(dst)

	if err != nil {
		return "", nil, err
	}
	m := http.DetectContentType(b[:512])

	return m, b, nil
}

func formatIntUnlimitedIf0(number int) string {
	if number == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", number)
}

type File struct {
	Name string `uri:"name" binding:"required"`
}

func StartServer() {
	var settings = GetSettings()

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "web/static")

	api := router.Group("/api")
	web := router.Group("/")

	files := api.Group("/files")

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tmpl", gin.H{})
	})

	files.POST("/", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		n, err := Upload(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		params := c.Request.URL.Query()
		if params.Get("redirect") == "true" {
			c.Redirect(http.StatusFound, fmt.Sprintf("%s/%s/", files.BasePath(), n))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": "Uploaded successfully",
			"name":    n,
		})
	})

	files.GET("/:name/", func(c *gin.Context) {
		var f File
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		m, cn, err := Download(f.Name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err})
			return
		}
		c.Header("Content-Disposition", "attachment; filename="+f.Name)
		c.Data(http.StatusOK, m, cn)
	})

	web.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":        settings.AppName,
			"filesize":     formatIntUnlimitedIf0(settings.FileSizeLimit),
			"persistance":  formatIntUnlimitedIf0(settings.FilePersistanceTime),
			"ratelimit":    formatIntUnlimitedIf0(settings.IPDayRateLimit),
			"storeLimit":   settings.IsStorePathSizeLimitEnabled(),
			"authRequired": settings.IsAuthEnabled(),
		})
	})

	log.Printf("Starting server on %s:%s", settings.Host, settings.Port)
	if err := router.Run(fmt.Sprintf("%s:%s", settings.Host, settings.Port)); err != nil {
		panic(err)
	}
}
