package web

import (
	"mime"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ConfigureEndpoints настраивает роутинг для веб приложения
func (s *service) ConfigureEndpoints() {
	routes := s.router.Group("", s.pagesController.ErrorPageMiddleware())

	routes.GET("/", s.pagesController.IndexPage)
	routes.GET("/search", s.pagesController.SearchPage)
	routes.GET("/bonds/:id", s.pagesController.BondPage)
	routes.GET("/collections/:id", s.pagesController.CollectionPage)
	routes.GET("/suggest", s.pagesController.SuggestPage)

	s.router.NoRoute(s.serveStaticFiles)
	mime.AddExtensionType(".js", "application/javascript")
}

// serveStaticFiles отвечает за раздачу статики
func (s *service) serveStaticFiles(c *gin.Context) {
	dir, file := path.Split(c.Request.RequestURI)
	ext := filepath.Ext(file)
	if file == "" || ext == "" {
		c.File("./www/index.html")
	} else {
		c.File("./www" + path.Join(dir, file))
	}
}
