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
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
)

// NGINXAdapter manages an NGINX based proxy.
type NGINXAdapter struct {
	client           nginx.Client
	manager          nginx.Manager
	discoveryMonitor monitor.DiscoveryMonitor
	rulesMonitor     monitor.RulesMonitor

	instances []api.ServiceInstance
	rules     []api.Rule
	mutex     sync.Mutex
}

// NewNGINXAdapter creates a new adapter instance.
func NewNGINXAdapter(conf *config.Config, discoveryMonitor monitor.DiscoveryMonitor, rulesMonitor monitor.RulesMonitor) *NGINXAdapter {
	client := nginx.NewClient("http://localhost:5813") // FIXME: hardcoded
	service := nginx.NewService(conf.Service.Name, conf.Service.Tags)
	manager := nginx.NewManager(nginx.Config{
		Client:  client,
		Service: service,
	})

	return &NGINXAdapter{
		client:           client,
		manager:          manager,
		discoveryMonitor: discoveryMonitor,
		rulesMonitor:     rulesMonitor,

		instances: []api.ServiceInstance{},
		rules:     []api.Rule{},
	}
}

// Start NGINX proxy.
func (a *NGINXAdapter) Start() error {
	a.discoveryMonitor.SetListeners([]monitor.DiscoveryListener{a})
	go func() {
		if err := a.discoveryMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Discovery monitor failed")
		}
	}()

	a.rulesMonitor.SetListeners([]monitor.RulesListener{a})
	go func() {
		if err := a.rulesMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Rules monitor failed")
		}
	}()

	return nil
}

// Stop NGINX proxy.
func (a *NGINXAdapter) Stop() error {
	if err := a.discoveryMonitor.Stop(); err != nil {
		return err
	}

	return a.rulesMonitor.Stop()
}

// CatalogChange updates on a change in the catalog.
func (a *NGINXAdapter) CatalogChange(instances []api.ServiceInstance) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.instances = instances
	return a.manager.Update(a.instances, a.rules)
}

// RuleChange updates NGINX on a change in the proxy configuration.
func (a *NGINXAdapter) RuleChange(rules []api.Rule) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.rules = rules
	return a.manager.Update(a.instances, a.rules)
}

// GetState returns the cached state of the NGINX adapter.
func (a *NGINXAdapter) GetState() ([]api.ServiceInstance, []api.Rule) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.instances, a.rules
}
