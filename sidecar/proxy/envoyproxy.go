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

package proxy

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/proxy/envoy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
)

// EnvoyProxy updates Envoy to reflect changes in the controller and registry
type EnvoyProxy interface {
	monitor.ControllerListener
	monitor.RegistryListener
	GetState() ([]api.ServiceInstance, []rules.Rule)
}

type envoyProxy struct {
	instances []api.ServiceInstance
	rules     []rules.Rule
	envoy     envoy.Manager
	mutex     sync.Mutex
}

// NewEnvoyProxy instantiates a new instance
func NewEnvoyProxy(envoyManager nginx.Manager) EnvoyProxy {
	return &envoyProxy{
		rules:     []rules.Rule{},
		instances: []api.ServiceInstance{},
		envoy:     envoyManager,
	}
}

// CatalogChange updates NGINX on a change in the catalog
func (p *envoyProxy) CatalogChange(instances []api.ServiceInstance) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.instances = instances
	return p.update()
}

// RuleChange updates NGINX on a change in the proxy configuration
func (p *envoyProxy) RuleChange(rules []rules.Rule) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.rules = rules
	return p.update()
}

func (p *envoyProxy) update() error {
	logrus.Debug("Updating Envoy")
	return p.envoy.Update(p.instances, p.rules)
}

func (p *envoyProxy) GetState() ([]api.ServiceInstance, []rules.Rule) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.instances, p.rules
}
