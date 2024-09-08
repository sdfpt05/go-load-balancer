package domain

import (
	"net/url"
	"sync/atomic"
	"time"
)

type Server struct {
	URL          *url.URL
	Active       atomic.Bool
	Connections  int64
	LastChecked  time.Time
	ResponseTime time.Duration
}
