package http

import (
	"context"
	"net/http"

	"example-api-template/internal/errs"
	"example-api-template/pkg/i18n"
	"example-api-template/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ------------------------
// I18n Middleware
// ------------------------

func I18nMiddleware(localizer *i18n.Localizer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := detectLanguage(c, localizer)
			ctx := localizer.SetLanguageInContext(c.Request().Context(), lang)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set("Content-Language", lang)
			return next(c)
		}
	}
}

func detectLanguage(c echo.Context, localizer *i18n.Localizer) string {
	// Priority: query -> header -> Accept-Language -> cookie -> default
	if lang := c.QueryParam("lang"); localizer.IsLanguageSupported(lang) {
		return lang
	}
	if lang := c.Request().Header.Get("X-Language"); localizer.IsLanguageSupported(lang) {
		return lang
	}
	if lang := localizer.ParseAcceptLanguage(c.Request().Header.Get("Accept-Language")); localizer.IsLanguageSupported(lang) {
		return lang
	}
	if cookie, err := c.Cookie("language"); err == nil && localizer.IsLanguageSupported(cookie.Value) {
		return cookie.Value
	}
	return localizer.DefaultLanguage()
}

// ------------------------
// CORS Middleware
// ------------------------

func CORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			setCORSHeaders(c)
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}
}

func setCORSHeaders(c echo.Context) {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Language, Accept-Language")
	c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Language, X-Total-Count")
	c.Response().Header().Set("Access-Control-Max-Age", "86400")
	if origin := c.Request().Header.Get("Origin"); origin != "" {
		c.Logger().Debugf("CORS request from origin: %s", origin)
	}
}

// ------------------------
// Request ID Middleware
// ------------------------

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := getRequestID(c)
			ctx := context.WithValue(c.Request().Context(), "request_id", requestID)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set("X-Request-ID", requestID)
			return next(c)
		}
	}
}

func getRequestID(c echo.Context) string {
	if id := c.Request().Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

// ------------------------
// Error Handler Middleware
// ------------------------

func ErrorHandlerMiddleware(localizer *i18n.Localizer) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		switch e := err.(type) {
		case *errs.AppError:
			handleAppError(e, c, localizer)
		case *echo.HTTPError:
			handleEchoError(e, c)
		default:
			logger.Debug("Unhandled error", zap.Any("error", err))
			sendErrorResponse(c, http.StatusInternalServerError, "Internal Server Error")
		}
	}
}

func handleAppError(appErr *errs.AppError, c echo.Context, localizer *i18n.Localizer) {
	ctx := c.Request().Context()
	localized := appErr.LocalizeWithContext(localizer, ctx)
	res := NewErrorResponse(string(localized.Code), localized.Err, localized.Message, localized.Details)

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			if err := c.NoContent(appErr.HTTPStatus); err != nil {
				c.Logger().Error(err)
			}
		} else {
			if err := c.JSON(appErr.HTTPStatus, res); err != nil {
				c.Logger().Error(err)
			}
		}
	}
}

func handleEchoError(he *echo.HTTPError, c echo.Context) {
	logger.Debug("Echo HTTPError detected", zap.Int("code", he.Code))
	sendErrorResponse(c, he.Code, he.Message)
}

// sendErrorResponse sends a generic error response
func sendErrorResponse(c echo.Context, code int, message interface{}) {
	c.Logger().Errorf("HTTP Error %d: %v", code, message)
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			if err := c.NoContent(code); err != nil {
				c.Logger().Error(err)
			}
		} else {
			if err := c.JSON(code, map[string]interface{}{"error": message}); err != nil {
				c.Logger().Error(err)
			}
		}
	}
}
