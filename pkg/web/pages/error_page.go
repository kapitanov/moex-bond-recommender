package pages

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ErrorPageModel - модель для страницы
type ErrorPageModel struct {
	StatusCode int
	Message    string
}

// Error - ошибка бизнес-логики
type Error struct {
	StatusCode int
	Message    string
}

// Error возвращает сообщение об ошибке
func (e Error) Error() string {
	return fmt.Sprintf("[HTTP %d] %s", e.StatusCode, e.Message)
}

// NewError создает новый объект типа Error
func NewError(statusCode int, message string, a ...interface{}) Error {
	message = fmt.Sprintf(message, a...)
	return Error{statusCode, message}
}

// ErrorPageMiddleware отвечает за обработку ошибок
func (ctrl *Controller) ErrorPageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			var model *ErrorPageModel
			switch e := err.(type) {
			case Error:
				model = &ErrorPageModel{
					StatusCode: e.StatusCode,
					Message:    e.Message,
				}
			default:
				ctrl.logger.Printf("error while handling \"%s %s\": %s", c.Request.Method, c.Request.RequestURI, err)
				model = &ErrorPageModel{
					StatusCode: 500,
					Message:    "Internal Server Error",
				}
			}

			ctrl.renderHTML(c, model.StatusCode, "pages/error", model)
		}()

		c.Next()
	}
}
