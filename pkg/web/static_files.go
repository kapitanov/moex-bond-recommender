package web

import (
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (s *service) ServeStaticFiles(c *gin.Context) {
	dir, file := path.Split(c.Request.RequestURI)
	ext := filepath.Ext(file)
	if file == "" || ext == "" {
		c.File("./www/index.html")
	} else {
		c.File("./www" + path.Join(dir, file))
	}
}
