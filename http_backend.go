// Copyright 2018 Adam Tauber
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fastcolly

import (
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/valyala/fasthttp"
)

type httpBackend struct {
	Client     *fasthttp.Client
	lock       *sync.RWMutex
}

// LimitRule provides connection restrictions for domains.
// Both DomainRegexp and DomainGlob can be used to specify
// the included domains patterns, but at least one is required.
// There can be two kind of limitations:
//  - Parallelism: Set limit for the number of concurrent requests to matching domains
//  - Delay: Wait specified amount of time between requests (parallelism is 1 in this case)
type LimitRule struct {
	// DomainRegexp is a regular expression to match against domains
	DomainRegexp string
	// DomainRegexp is a glob pattern to match against domains
	DomainGlob string
	// Delay is the duration to wait before creating a new request to the matching domains
	Delay time.Duration
	// RandomDelay is the extra randomized duration to wait added to Delay before creating a new request
	RandomDelay time.Duration
	// Parallelism is the number of the maximum allowed concurrent requests of the matching domains
	Parallelism    int
	waitChan       chan bool
	compiledRegexp *regexp.Regexp
	compiledGlob   glob.Glob
}

// Init initializes the private members of LimitRule
func (r *LimitRule) Init() error {
	waitChanSize := 1
	if r.Parallelism > 1 {
		waitChanSize = r.Parallelism
	}
	r.waitChan = make(chan bool, waitChanSize)
	hasPattern := false
	if r.DomainRegexp != "" {
		c, err := regexp.Compile(r.DomainRegexp)
		if err != nil {
			return err
		}
		r.compiledRegexp = c
		hasPattern = true
	}
	if r.DomainGlob != "" {
		c, err := glob.Compile(r.DomainGlob)
		if err != nil {
			return err
		}
		r.compiledGlob = c
		hasPattern = true
	}
	if !hasPattern {
		return ErrNoPattern
	}
	return nil
}

func (h *httpBackend) Init(jar http.CookieJar) {
	rand.Seed(time.Now().UnixNano())
	h.Client = &fasthttp.Client{

	}

	//h.Client = &http.Client{
	//	Jar:     jar,
	//	Timeout: 10 * time.Second,
	//}

	h.lock = &sync.RWMutex{}
}

// Match checks that the domain parameter triggers the rule
func (r *LimitRule) Match(domain string) bool {
	match := false
	if r.compiledRegexp != nil && r.compiledRegexp.MatchString(domain) {
		match = true
	}
	if r.compiledGlob != nil && r.compiledGlob.Match(domain) {
		match = true
	}
	return match
}

//func (h *httpBackend) GetMatchingRule(domain string) *LimitRule {
//	if h.LimitRules == nil {
//		return nil
//	}
//	h.lock.RLock()
//	defer h.lock.RUnlock()
//	for _, r := range h.LimitRules {
//		if r.Match(domain) {
//			return r
//		}
//	}
//	return nil
//}

func (h *httpBackend) Cache(request *fasthttp.Request,res *fasthttp.Response, bodySize int, cacheDir string) (*Response, error) {
	if cacheDir == "" || string(request.Header.Method()) != "GET" {
		return h.Do(request,res)
	}
	sum := sha1.Sum(request.RequestURI())
	hash := hex.EncodeToString(sum[:])
	dir := path.Join(cacheDir, hash[:2])
	filename := path.Join(dir, hash)
	if file, err := os.Open(filename); err == nil {
		resp := new(Response)
		err := gob.NewDecoder(file).Decode(resp)
		file.Close()
		if resp.StatusCode < 500 {
			return resp, err
		}
	}
	resp, err := h.Do(request,res)
	if err != nil || resp.StatusCode >= 500 {
		return resp, err
	}
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return resp, err
		}
	}
	file, err := os.Create(filename + "~")
	if err != nil {
		return resp, err
	}
	if err := gob.NewEncoder(file).Encode(resp); err != nil {
		file.Close()
		return resp, err
	}
	file.Close()
	return resp, os.Rename(filename+"~", filename)
}

func (h *httpBackend) Do(req *fasthttp.Request,res *fasthttp.Response) (*Response, error) {

	if err := h.Client.Do(req,res);err!=nil {
		return nil,err
	}

	rhdr := make(http.Header)
	res.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		rhdr.Set(sk, sv)
	})

	return &Response{
		StatusCode: res.StatusCode(),
		Body:       res.Body(),
		Headers:    &rhdr,
	}, nil
}

//func (h *httpBackend) Limit(rule *LimitRule) error {
//	h.lock.Lock()
//	if h.LimitRules == nil {
//		h.LimitRules = make([]*LimitRule, 0, 8)
//	}
//	h.LimitRules = append(h.LimitRules, rule)
//	h.lock.Unlock()
//	return rule.Init()
//}
//
//func (h *httpBackend) Limits(rules []*LimitRule) error {
//	for _, r := range rules {
//		if err := h.Limit(r); err != nil {
//			return err
//		}
//	}
//	return nil
//}
