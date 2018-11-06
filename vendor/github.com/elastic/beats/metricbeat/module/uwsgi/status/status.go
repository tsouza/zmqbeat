// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package status

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
	"github.com/elastic/beats/metricbeat/module/uwsgi"
)

func init() {
	mb.Registry.MustAddMetricSet("uwsgi", "status", New,
		mb.WithHostParser(uwsgi.HostParser),
		mb.DefaultMetricSet(),
	)
}

// MetricSet for fetching uwsgi metrics from StatServer.
type MetricSet struct {
	mb.BaseMetricSet
}

// New creates a new instance of the MetricSet.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Beta("The uWSGI status metricset is beta")
	return &MetricSet{BaseMetricSet: base}, nil
}

func fetchStatData(URL string) ([]byte, error) {
	var reader io.Reader

	u, err := url.Parse(URL)
	if err != nil {
		logp.Err("parsing uwsgi stats url failed: ", err)
		return nil, err
	}

	switch u.Scheme {
	case "tcp":
		conn, err := net.Dial(u.Scheme, u.Host)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		reader = conn
	case "unix":
		path := strings.Replace(URL, "unix://", "", -1)
		conn, err := net.Dial(u.Scheme, path)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		reader = conn
	case "http", "https":
		res, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			logp.Err("failed to fetch uwsgi status with code: ", res.StatusCode)
			return nil, errors.New("http failed")
		}
		reader = res.Body
	default:
		return nil, errors.New("unknown scheme")
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		logp.Err("uwsgi data read failed: ", err)
		return nil, err
	}

	return data, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	content, err := fetchStatData(m.HostData().URI)
	if err != nil {
		return []common.MapStr{
			common.MapStr{
				"error": err.Error(),
			},
		}, err
	}
	return eventsMapping(content)
}
