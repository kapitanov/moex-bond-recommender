package pages

import (
	"bytes"
	"log"
	"net/http"

	"github.com/foolin/goview"
	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
)

// Controller - контроллер веб-страниц
type Controller struct {
	app        app.App
	viewEngine *goview.ViewEngine
	logger     *log.Logger
}

// New создает новый Controller
func New(app app.App, googleAnalyticsID string, debugMode bool, logger *log.Logger) *Controller {
	viewEngine := goview.New(goview.Config{
		Root:         "templates",
		Extension:    ".html",
		Master:       "layout",
		Partials:     []string{},
		Funcs:        DefineFunctions(googleAnalyticsID),
		DisableCache: debugMode,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	})

	return &Controller{app: app, viewEngine: viewEngine, logger: logger}
}

func (ctrl *Controller) renderHTML(c *gin.Context, status int, template string, model interface{}) {
	w := newBufferedHttpWriter()

	err := ctrl.viewEngine.Render(w, status, template, model)
	if err != nil {
		panic(err)
	}

	err = w.Flush(c.Writer)
	if err != nil {
		panic(err)
	}
}

type bufferedHttpWriter struct {
	statusCode int
	header     http.Header
	buffer     bytes.Buffer
}

func newBufferedHttpWriter() *bufferedHttpWriter {
	return &bufferedHttpWriter{
		buffer: bytes.Buffer{},
		header: make(http.Header),
	}
}

func (w *bufferedHttpWriter) Header() http.Header {
	return w.header
}

func (w *bufferedHttpWriter) Write(buf []byte) (int, error) {
	return w.buffer.Write(buf)
}

func (w *bufferedHttpWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bufferedHttpWriter) Flush(wr http.ResponseWriter) error {
	wr.WriteHeader(w.statusCode)

	header := wr.Header()
	for key, values := range w.header {
		for _, value := range values {
			header.Add(key, value)
		}
	}

	_, err := wr.Write(w.buffer.Bytes())
	return err
}
