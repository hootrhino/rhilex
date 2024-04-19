// Copyright (C) 2024 wwhai
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
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hootrhino/rhilex/typex"
)

type user struct {
	ts     int64
	tokens uint
}

func clearInBackground(data *sync.Map, rate int64) {
	for {
		select {
		case <-typex.GCTX.Done():
			{
				return
			}
		default:
			{
			}
		}
		data.Range(func(k, v interface{}) bool {
			if v.(user).ts+rate <= time.Now().Unix() {
				data.Delete(k)
			}
			return true
		})
		time.Sleep(time.Minute)
	}
}

type inMemoryStoreType struct {
	rate  int64
	limit uint
	data  *sync.Map
	skip  func(ctx *gin.Context) bool
}

func (s *inMemoryStoreType) Limit(key string, c *gin.Context) Info {
	var u user
	m, ok := s.data.Load(key)
	if !ok {
		u = user{time.Now().Unix(), s.limit}
	} else {
		u = m.(user)
	}
	if u.ts+s.rate <= time.Now().Unix() {
		u.tokens = s.limit
	}
	if s.skip != nil && s.skip(c) {
		return Info{
			Limit:         s.limit,
			RateLimited:   false,
			ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
			RemainingHits: u.tokens,
		}
	}
	if u.tokens <= 0 {
		return Info{
			Limit:         s.limit,
			RateLimited:   true,
			ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
			RemainingHits: 0,
		}
	}
	u.tokens--
	u.ts = time.Now().Unix()
	s.data.Store(key, u)
	return Info{
		Limit:         s.limit,
		RateLimited:   false,
		ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * time.Second.Nanoseconds())),
		RemainingHits: u.tokens,
	}
}

type InMemoryOptions struct {
	// the user can make Limit amount of requests every Rate
	Rate time.Duration
	// the amount of requests that can be made every Rate
	Limit uint
	// a function that returns true if the request should not count toward the rate limit
	Skip func(*gin.Context) bool
}

func InMemoryStore(options *InMemoryOptions) Store {
	data := &sync.Map{}
	store := inMemoryStoreType{int64(options.Rate.Seconds()), options.Limit, data, options.Skip}
	go clearInBackground(data, store.rate)
	return &store
}
