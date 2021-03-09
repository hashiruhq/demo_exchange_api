package logger

import (
	"net/http"
	"regexp"
	"time"

	"github.com/rs/xid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config for logger
type Config struct {
	Logger *zerolog.Logger
	// UTC a boolean stating whether to use UTC time zone or local.
	UTC            bool
	SkipPath       []string
	SkipPathRegexp *regexp.Regexp
}

// GetLogger from gin context
func GetLogger(c *gin.Context) zerolog.Logger {
	if logger, ok := c.Get("_log"); ok {
		return logger.(zerolog.Logger)
	}
	return log.Logger
}

// SetLogger initializes the logging middleware.
func SetLogger(config ...Config) gin.HandlerFunc {
	var newConfig Config
	if len(config) > 0 {
		newConfig = config[0]
	}
	var skip map[string]struct{}
	if length := len(newConfig.SkipPath); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range newConfig.SkipPath {
			skip[path] = struct{}{}
		}
	}

	var sublog zerolog.Logger
	if newConfig.Logger == nil {
		sublog = log.Logger
	} else {
		sublog = *newConfig.Logger
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		track := true

		if c.Request.Method == "GET" {
			track = false
		}

		if _, ok := skip[path]; ok {
			track = false
		}

		if track &&
			newConfig.SkipPathRegexp != nil &&
			newConfig.SkipPathRegexp.MatchString(path) {
			track = false
		}
		
		id := xid.New().String()
		c.Writer.Header().Set("X-Request-Id", id)
		var reqlogger zerolog.Logger
		if track {
			reqlogger = sublog.With().
				Str("request_id", id).
				Logger()

			c.Set("_log", reqlogger)
		}
		c.Next()
		
		if track {
			end := time.Now()
			latency := end.Sub(start)
			if newConfig.UTC {
				end = end.UTC()
			}

			msg := "Request"
			if len(c.Errors) > 0 {
				msg = c.Errors.String()
			}

			dumplogger := reqlogger.With().
				Str("method", c.Request.Method).
				Str("path", path).
				Str("ip", c.ClientIP()).
				Str("user-agent", c.Request.UserAgent()).
				Int("status", c.Writer.Status()).
				Dur("latency", latency).
				Logger()

			switch {
			case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
				dumplogger.Warn().Msg(msg)
			case c.Writer.Status() >= http.StatusInternalServerError:
				dumplogger.Error().Msg(msg)
			default: 
				dumplogger.Info().Msg(msg)
			}
		}

	}
}