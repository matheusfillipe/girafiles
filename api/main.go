package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

type File struct {
	Name string `uri:"name" binding:"required"`
}

func StartServer() {
	var settings = GetSettings()
	GetDB().createTable()

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "web/static")

	api := router.Group("/api")
	web := router.Group("/")
	files := router.Group("/files")

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
		// Handle storage size after uploading a file
		c.Next()
		db := GetDB()
		log.Printf("Checking what we can delete")

		// Delete expired files
		namesToDelete, err := db.deleteExpiredFiles()
		if err != nil {
			log.Printf("Error deleting expired files: %s", err)
		}
		for _, name := range namesToDelete {
			err := os.Remove(filepath.Join(settings.StorePath, name))
			if err != nil {
				log.Printf("Error deleting file %s: %s", name, err)
			}
			log.Printf("Deleted file %s because it expired", name)
		}

		// Delete oldest file if storage limit is exceeded
		if isStorageLimitExceeded() {
			namesToDelete, err := db.deleteOldestFiles(1)
			if err != nil {
				log.Printf("Error deleting oldest files from database: %s", err)
			}
			for _, name := range namesToDelete {
				err := os.Remove(filepath.Join(settings.StorePath, name))
				if err != nil {
					log.Printf("Error deleting file %s: %s", name, err)
				}
				log.Printf("Deleted file %s because storage limit was exceeded", name)
			}
		}
	})

	api.POST("/files", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := c.Request.URL.Query()
		n, err := Upload(file, c.ClientIP())
		if err != nil {
			if err.Error() == DUP_ENTRY_ERROR {
				if params.Get("redirect") == "true" {
					c.Redirect(http.StatusFound, fmt.Sprintf("%s/%s/", files.BasePath(), n))
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "File already exists",
					"url":     fmt.Sprintf("%s/%s", getHostUrl(c.Request)+files.BasePath(), n),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if params.Get("redirect") == "true" {
			c.Redirect(http.StatusFound, fmt.Sprintf("%s/%s", files.BasePath(), n))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"url":    fmt.Sprintf("%s/%s", getHostUrl(c.Request)+files.BasePath(), n),
		})
	})

	files.GET("/:name", func(c *gin.Context) {
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
		// If mime type is supported to be displayed in the browser, display it.
		// otherwise, download it.
		if isSupportedMimetype(m) {
			c.Data(http.StatusOK, m, cn)
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
			"uploadEP":     getHostUrl(c.Request) + api.BasePath() + "/files/",
		})
	})

	log.Printf("Starting server on %s:%s", settings.Host, settings.Port)
	if err := router.Run(fmt.Sprintf("%s:%s", settings.Host, settings.Port)); err != nil {
		panic(err)
	}
}
