// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package eureka

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"

	eurekaapi "github.com/amalgam8/amalgam8/registry/server/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	requestTimeout = time.Duration(60) * time.Second
)

type client struct {
	httpClient *http.Client
	eurekaURLs []string

	logger *log.Entry
}

func newClient(eurekaURLs []string) (*client, error) {
	var httpsRequired bool
	logger := logging.GetLogger(module)

	urls := make([]string, len(eurekaURLs))
	for i, eu := range eurekaURLs {
		for strings.HasSuffix(eu, "/") {
			eu = strings.TrimSuffix(eu, "/")
		}

		u, err := url.Parse(eu)
		if err != nil {
			return nil, err
		}

		if u.Scheme == "https" {
			httpsRequired = true
		}

		urls[i] = eu
	}

	hc := &http.Client{
		Timeout: requestTimeout,
	}

	if httpsRequired {
		hc.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &client{
		httpClient: hc,
		eurekaURLs: urls,
		logger:     logger,
	}, nil
}

func (c *client) getApplications(path string) (*eurekaapi.Applications, error) {
	var err error

	for _, eurl := range c.eurekaURLs {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", eurl, path), nil)
		req.Header.Set("Accept", "application/json")

		resp, err2 := c.httpClient.Do(req)
		if err2 != nil {
			err = err2
			continue
		}

		defer resp.Body.Close()

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			err = err2
			continue
		}

		var appsList eurekaapi.ApplicationsList
		err2 = json.Unmarshal(body, &appsList)
		if err2 != nil {
			err = err2
			continue
		}

		return appsList.Applications, nil
	}

	return nil, err
}

func (c *client) getApplicationsFull() (*eurekaapi.Applications, error) {
	return c.getApplications("apps")
}

func (c *client) getApplicationsDelta() (*eurekaapi.Applications, error) {
	apps, err := c.getApplications("apps/delta")
	if err != nil {
		return nil, err
	}

	if apps == nil || apps.VersionDelta == -1 {
		return nil, fmt.Errorf("Delta is not supported")
	}

	return apps, nil
}
