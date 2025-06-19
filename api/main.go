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

const (
	CONTENT_TYPE_JSON = "application/json"
	CONTENT_TYPE_TEXT = "text/plain"
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

func handleUpload(c *gin.Context, filename string, err error, params url.Values, contentType string) {
	url := fmt.Sprintf("%s/%s", getHostUrl(c.Request), filename)
	if err != nil {
		if err.Error() == DUP_ENTRY_ERROR {
			if params.Get("redirect") == "true" {
				slog.Debug(fmt.Sprintf("Redirecting to %s", url))
				c.Redirect(http.StatusFound, url)
				return
			}
			if contentType == CONTENT_TYPE_JSON {
				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "File already exists",
					"url":     url,
				})
			} else {
				c.String(http.StatusOK, "File already exists: %s", url)
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if params.Get("redirect") == "true" {
		slog.Debug(fmt.Sprintf("Redirecting to %s", url))
		c.Redirect(http.StatusFound, url)
		return
	}

	if contentType == CONTENT_TYPE_JSON {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"url":    url,
		})
	} else {
		c.String(http.StatusOK, url)
	}
}

func deliverFile(c *gin.Context, err error, file fileResponse, download bool) {
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no rows in result") {
			c.HTML(http.StatusNotFound, "404.tmpl", gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	// If mime type is supported to be displayed in the browser, display it.
	// otherwise, download it.
	if isSupportedMimetype(file.mimetype) && !download {
		c.Data(http.StatusOK, file.mimetype, file.content)
		return
	} else if download {
		c.Header("Content-Disposition", "attachment; filename="+file.name)
		c.Data(http.StatusOK, file.mimetype, file.content)
	} else {
		c.Redirect(308, fmt.Sprintf("/info/%s", file.shortname))
	}
}
func postFile(contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := c.Request.URL.Query()
		n, err := Upload(file, c.ClientIP())
		handleUpload(c, n, err, params, contentType)
	}
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
	if settings.TrustedProxyIP != "" {
		if err := router.SetTrustedProxies([]string{settings.TrustedProxyIP}); err != nil {
			panic(err)
		}
		router.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-Ip"}
		router.TrustedPlatform = settings.TrustedProxyIP
		router.ForwardedByClientIP = true
	}

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
		// Log the request headers
		fmt.Printf("Request Headers: %v", c.Request.Header)
		if settings.IsAuthEnabled() {
			checkAuth(c)
		}
		c.Next()
		go func() {
			<-c.Request.Context().Done()
			cleanup()
		}()
	})

	api.POST("/", postFile(CONTENT_TYPE_JSON))
	router.POST("/", postFile(CONTENT_TYPE_TEXT))

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
		handleUpload(c, n, err, params, CONTENT_TYPE_JSON)
	})
	files.GET("/info/:name", func(c *gin.Context) {
		var f File
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		file, err := Download(f.Name)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.tmpl", gin.H{})
		}
		c.HTML(http.StatusOK, "info.tmpl", gin.H{
			"title": settings.AppName,
			"size":  humanReadableSize(len(file.content)),
		})
	})
	files.GET("/:name", func(c *gin.Context) {
		var f File
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		file, err := Download(f.Name)
		deliverFile(c, err, file, c.Request.URL.Query().Get("download") != "")
	})

	files.GET("/:name/p", func(c *gin.Context) {
		var f File
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		file, err := Download(f.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		params := c.Request.URL.Query()
		lang := ""
		for _, param := range []string{"language", "lang", "l"} {
			_lang := params.Get(param)
			if isSupportedLanguage(_lang) {
				lang = "language-" + _lang
				break
			}
		}
		c.HTML(http.StatusOK, "paste.tmpl", gin.H{
			"title":          settings.AppName,
			"filename":       file.name,
			"class":          lang,
			"code":           string(file.content),
			"pasteLanguages": LANGUAGE_NAMES_MAP,
		})
	})

	files.GET("/:name/:alias", func(c *gin.Context) {
		var fb FileBucket
		if err := c.ShouldBindUri(&fb); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		file, err := DownloadFromBucket(fb.Bucket, fb.Name)
		deliverFile(c, err, file, c.Request.URL.Query().Get("download") != "")
	})

	files.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":          settings.AppName,
			"filesize":       formatIntUnlimitedIf0(settings.FileSizeLimit),
			"persistance":    formatIntUnlimitedIf0(settings.FilePersistanceTime),
			"ratelimit":      formatIntUnlimitedIf0(settings.IPDayRateLimit),
			"storeLimit":     settings.IsStorePathSizeLimitEnabled(),
			"authRequired":   settings.IsAuthEnabled(),
			"uploadEP":       getHostUrl(c.Request),
			"pasteLanguages": LANGUAGE_NAMES_MAP,
		})
	})

	log.Printf("Starting server on %s:%s", settings.Host, settings.Port)
	if err := router.Run(fmt.Sprintf("%s:%s", settings.Host, settings.Port)); err != nil {
		panic(err)
	}
}
