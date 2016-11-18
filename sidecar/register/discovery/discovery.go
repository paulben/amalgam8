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

package discovery

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/api"
)

const (
	routeParamServiceName = "sname"
	apiVer                = "/v1"
	registrationPath      = apiVer + "/registration"
	registrationTemplate  = registrationPath + "/#" + routeParamServiceName
)

// Hosts is the array of hosts returned by the GET registration
type Hosts struct {
	Hosts []Host `json:"hosts"`
}

// Host is the endpoint and tag data for a service instance
type Host struct {
	IPAddr string            `json:"ip_address"`
	Port   uint16            `json:"port"`
	Tags   map[string]string `json:"tags"`
}

// Discovery handles discovery API calls
type Discovery struct {
	discovery api.ServiceDiscovery
}

// NewDiscovery creates struct
func NewDiscovery(discovery api.ServiceDiscovery) *Discovery {
	return &Discovery{
		discovery: discovery,
	}
}

// Routes for discovery API
func (d *Discovery) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Get(registrationTemplate, d.getRegistration),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// getRegistration
func (d *Discovery) getRegistration(w rest.ResponseWriter, req *rest.Request) {
	sname := req.PathParam(routeParamServiceName)
	instances, err := d.discovery.ListServiceInstances(sname)
	if err != nil {
		logrus.WithError(err).Warnf("Failed to get the list of service instances")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := Hosts{
		Hosts: translate(instances),
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&resp)
	return
}

func translate(instances []*api.ServiceInstance) []Host {
	hosts := []Host{}
	tags := make(map[string]string)

	for _, instance := range instances {
		var host Host
		ip, port, err := splitHostPort(instance.Endpoint)
		if err != nil {
			logrus.WithError(err).Warnf("unable to resolve ip address for instance '%s'", instance.ID)
			continue
		}
		host = Host{IPAddr: ip, Port: port, Tags: tags}
		hosts = append(hosts, host)
	}

	return hosts
}

func splitHostPort(endpoint api.ServiceEndpoint) (string, uint16, error) {
	switch endpoint.Type {
	case "tcp", "udp":
		return splitHostPortTCPUDP(endpoint.Value)
	case "http", "https":
		return splitHostPortHTTP(endpoint.Value)
	default:
		return "", 0, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
	}
}

func splitHostPortTCPUDP(value string) (string, uint16, error) {
	// Assume value is "host:port"
	host, port, err := net.SplitHostPort(value)

	// Assume value is "host" (no port)
	if err != nil {
		host = value
		port = "0"
	}

	// A url will wind up being filtered out - should it be included?
	ip := net.ParseIP(host)
	if ip == nil {
		return "", 0, fmt.Errorf("could not parse '%s' as ip:port", value)
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}

	return ip.String(), uint16(portNum), nil
}

func splitHostPortHTTP(value string) (string, uint16, error) {
	isHTTP := strings.HasPrefix(value, "http://")
	isHTTPS := strings.HasPrefix(value, "https://")
	if !isHTTPS && !isHTTP {
		value = "http://" + value
		isHTTP = true
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		return "", 0, err
	}

	ip, port, err := splitHostPortTCPUDP(parsedURL.Host)
	if err != nil {
		return "", 0, err
	}

	// Use default port, if not specified
	if port == 0 {
		if isHTTP {
			port = 80
		} else if isHTTPS {
			port = 443
		}
	}

	return ip, port, nil
}
