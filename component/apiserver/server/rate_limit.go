// Copyright (C) 2023 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type Info struct {
	Limit         uint
	RateLimited   bool
	ResetTime     time.Time
	RemainingHits uint
}

type Store interface {
	// Limit takes in a key and *gin.Context and should return whether that key is allowed to make another request
	Limit(key string, c *gin.Context) Info
}

type Options struct {
	ErrorHandler func(*gin.Context, Info)
	KeyFunc      func(*gin.Context) string
	// a function that lets you check the rate limiting info and modify the response
	BeforeResponse func(c *gin.Context, info Info)
}

// RateLimiter is a function to get gin.HandlerFunc
func RateLimiter(s Store, options *Options) gin.HandlerFunc {
	if options == nil {
		options = &Options{}
	}
	if options.ErrorHandler == nil {
		options.ErrorHandler = func(c *gin.Context, info Info) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
			c.String(429, "Too many requests")
		}
	}
	if options.BeforeResponse == nil {
		options.BeforeResponse = func(c *gin.Context, info Info) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%v", info.RemainingHits))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
		}
	}
	if options.KeyFunc == nil {
		options.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP() + c.FullPath()
		}
	}
	return func(c *gin.Context) {
		key := options.KeyFunc(c)
		info := s.Limit(key, c)
		options.BeforeResponse(c, info)
		if c.IsAborted() {
			return
		}
		if info.RateLimited {
			options.ErrorHandler(c, info)
			c.Abort()
		} else {
			c.Next()
		}
	}
}
