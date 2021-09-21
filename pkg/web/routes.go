package web

import "mime"

// ConfigureEndpoints настраивает роутинг для веб приложения
func (s *service) ConfigureEndpoints() {
	s.router.GET("/", s.pagesController.IndexPage)
	s.router.GET("/search", s.pagesController.SearchPage)
	s.router.GET("/bonds/:id", s.pagesController.BondPage)
	s.router.GET("/collections/:id", s.pagesController.CollectionPage)

	s.router.NoRoute(s.ServeStaticFiles)
	mime.AddExtensionType(".js", "application/javascript")
}
