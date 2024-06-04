package api

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func formatIntUnlimitedIf0(number int) string {
	if number == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", number)
}

// Get the host URL either from this server or from x-forwarded-host and x-forwarded-proto headers
// if they are available.
func getHostUrl(request *http.Request) string {
	if request.Header.Get("x-forwarded-host") != "" && request.Header.Get("x-forwarded-proto") != "" {
		return fmt.Sprintf("%s://%s", request.Header.Get("x-forwarded-proto"), request.Header.Get("x-forwarded-host"))
	}
	proto := "http"
	if request.TLS != nil {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s", proto, request.Host)
}

func checkAuth(c *gin.Context) bool {
	users := GetSettings().Users
	gin.BasicAuth(gin.Accounts(users))(c)
	if c.IsAborted() {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}
	return true
}

type FileParams interface {
	GetName() string
}

type File struct {
	Name string `uri:"name" binding:"required"`
}

func (f File) GetName() string {
	return f.Name
}

type FileBucket struct {
	Bucket string `uri:"name" binding:"required"`
	Name   string `uri:"alias" binding:"required"`
}

// GetName implements FileRequest.
func (fb FileBucket) GetName() string {
	return fb.Name
}

func handleUpload(c *gin.Context, filename string, err error, params url.Values, files *gin.RouterGroup) {
	if err != nil {
		if err.Error() == DUP_ENTRY_ERROR {
			if params.Get("redirect") == "true" {
				c.Redirect(http.StatusFound, filename)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "File already exists",
				"url":     fmt.Sprintf("%s/%s", getHostUrl(c.Request), filename),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if params.Get("redirect") == "true" {
		c.Redirect(http.StatusFound, filename)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"url":    fmt.Sprintf("%s/%s", getHostUrl(c.Request), filename),
	})
}

func deliverFile(c *gin.Context, err error, f FileParams, m string, cn []byte) {
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no rows in result") {
			c.HTML(http.StatusNotFound, "404.tmpl", gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	// If mime type is supported to be displayed in the browser, display it.
	// otherwise, download it.
	if isSupportedMimetype(m) {
		c.Data(http.StatusOK, m, cn)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+f.GetName())
	c.Data(http.StatusOK, m, cn)
}

func StartServer() {
	var settings = GetSettings()

	slog.Debug("Creating database Tables...")
	GetDB().createTable()
	slog.Debug("Done.")

	router := gin.Default()
	router.RemoveExtraSlash = true
	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "web/static")

	api := router.Group("/api")
	files := router.Group("/")

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, api.BasePath()) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.HTML(http.StatusNotFound, "404.tmpl", gin.H{})
	})

	api.Use(func(c *gin.Context) {
		if settings.IsAuthEnabled() {
			checkAuth(c)
		}
		c.Next()
		go func() {
			<-c.Request.Context().Done()
			cleanup()
		}()
	})

	api.POST("/", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := c.Request.URL.Query()
		n, err := Upload(file, c.ClientIP())
		handleUpload(c, n, err, params, files)
	})

	api.PUT("/:name/:alias", func(c *gin.Context) {
		var fb FileBucket
		if err := c.ShouldBindUri(&fb); err != nil {
			slog.Error(fmt.Sprintf("Failed to upload file: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if len(fb.Bucket) > 64 || len(fb.Bucket) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bucket name must be between 4 and 64 characters"})
			return
		}

		// Receive file as request content
		reader := c.Request.Body
		params := c.Request.URL.Query()
		n, err := UploadToBucket(reader, c.ClientIP(), fb.Bucket, fb.Name)
		handleUpload(c, n, err, params, files)
	})

	files.GET("/:name", func(c *gin.Context) {
		var f File
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		m, cn, err := Download(f.Name)
		deliverFile(c, err, f, m, cn)
	})

	files.GET("/:name/:alias", func(c *gin.Context) {
		var fb FileBucket
		if err := c.ShouldBindUri(&fb); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		m, cn, err := DownloadFromBucket(fb.Bucket, fb.Name)
		deliverFile(c, err, fb, m, cn)
	})

	files.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":        settings.AppName,
			"filesize":     formatIntUnlimitedIf0(settings.FileSizeLimit),
			"persistance":  formatIntUnlimitedIf0(settings.FilePersistanceTime),
			"ratelimit":    formatIntUnlimitedIf0(settings.IPDayRateLimit),
			"storeLimit":   settings.IsStorePathSizeLimitEnabled(),
			"authRequired": settings.IsAuthEnabled(),
			"uploadEP":     getHostUrl(c.Request) + api.BasePath() + "/",
		})
	})

	log.Printf("Starting server on %s:%s", settings.Host, settings.Port)
	if err := router.Run(fmt.Sprintf("%s:%s", settings.Host, settings.Port)); err != nil {
		panic(err)
	}
}
