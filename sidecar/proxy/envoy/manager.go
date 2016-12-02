package envoy

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sort"

	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/api"
)

// Manager for updating envoy
type Manager interface {
	Update(instances []api.ServiceInstance, rules []rules.Rule) error
}

// NewManager creates new instance
func NewManager(serviceName string) Manager {
	return &manager{
		serviceName: serviceName,
		service:     NewService("/etc/envoy.json"),
	}
}

type manager struct {
	serviceName string
	service     Service
}

func (m *manager) Update(instances []api.ServiceInstance, rules []rules.Rule) error {
	conf, err := generateConfig(rules, instances, m.serviceName)
	if err != nil {
		return err
	}

	file, err := os.Create("/etc/envoy.json")
	if err != nil {
		return err
	}

	if err := writeConfig(file, conf); err != nil {
		file.Close()
		return err
	}
	file.Close()

	if err := m.service.Reload(); err != nil {
		return err
	}

	return nil
}

func writeConfig(writer io.Writer, root Root) error {
	out, err := json.MarshalIndent(&root, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(out)
	return err
}

func ParseServiceName(s string) (string, []string) {
	var res []string
	buf := bytes.NewBuffer([]byte{})
	b := []byte(s)

	i := 0
	for i = 0; i < len(b)-1; i++ {
		if b[i] == '_' {
			switch b[i+1] {
			case '_':
				buf.WriteByte('_')
			case 's':
				res = append(res, buf.String())
				buf = bytes.NewBuffer([]byte{})
			default:
				logrus.WithField("character", b[i+1]).Warn("Unrecognized control character")
			}
			i++
		} else {
			buf.WriteByte(b[i])
		}
	}

	if i < len(b) {
		buf.WriteByte(b[i])
	}
	res = append(res, buf.String())

	service := res[0]
	tags := res[1:]

	return service, tags
}

func BuildServiceName(s string, tags []string) string {
	sort.Strings(tags) // FIXME: by reference

	f := func(s string, buf *bytes.Buffer) {
		b := []byte(s)
		for i := range b {
			if b[i] == '_' {
				buf.WriteByte('_')
			}
			buf.WriteByte(b[i])
		}
	}

	buf := bytes.NewBuffer([]byte{})
	f(s, buf)
	for i := range tags {
		buf.WriteString("_s")
		f(tags[i], buf)
	}

	return buf.String()
}

func generateConfig(rules []rules.Rule, instances []api.ServiceInstance, serviceName string) (Root, error) {
	sanitizeRules(rules)
	rules = addDefaultRouteRules(rules, instances)

	clusters, err := buildClusters(rules, instances)
	if err != nil {
		return Root{}, err
	}

	routes := buildRoutes(rules)

	filters, err := buildFaults(rules, serviceName)
	if err != nil {
		return Root{}, err
	}

	if err := buildFS(rules); err != nil {
		return Root{}, err
	}

	return Root{
		RootRuntime: RootRuntime{
			SymlinkRoot:  RuntimePath,
			Subdirectory: "traffic_shift",
		},
		Listeners: []Listener{
			{
				Port: 6379,
				Filters: []NetworkFilter{
					{
						Type: "read",
						Name: "http_connection_manager",
						Config: NetworkFilterConfig{
							CodecType:  "auto",
							StatPrefix: "ingress_http",
							RouteConfig: RouteConfig{
								VirtualHosts: []VirtualHost{
									{
										Name:    "backend",
										Domains: []string{"*"},
										Routes:  routes,
									},
								},
							},
							Filters: filters,
						},
					},
				},
			},
		},
		Admin: Admin{
			AccessLogPath: "/var/log/envoy_access.log",
			Port:          8001,
		},
		ClusterManager: ClusterManager{
			Clusters: clusters,
			SDS: SDS{
				Cluster: Cluster{
					Name:             "sds",
					Type:             "strict_dns",
					ConnectTimeoutMs: 1000,
					LbType:           "round_robin",
					Hosts: []Host{
						{
							URL: "tcp://127.0.0.1:6500",
						},
					},
				},
				RefreshDelayMs: 1000,
			},
		},
	}, nil
}

type uniqueRoute struct {
	Service string
	Tags    []string
}

func buildClusters(rules []rules.Rule, instances []api.ServiceInstance) ([]Cluster, error) {
	// Find unique routes
	uniqueRoutes := make(map[string]uniqueRoute)
	for _, rule := range rules {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				key := BuildServiceName(backend.Name, backend.Tags)

				uniqueRoutes[key] = uniqueRoute{
					Service: backend.Name,
					Tags:    backend.Tags,
				}
			}
		}
	}

	clusterMap := make(map[string]struct{})
	for _, instance := range instances {
		instanceTags := make(map[string]struct{})
		for _, tag := range instance.Tags {
			instanceTags[tag] = struct{}{}
		}

		for key, uniqueRoute := range uniqueRoutes {
			if uniqueRoute.Service == instance.ServiceName {
				isSubset := true
				for _, tag := range uniqueRoute.Tags {
					if _, exists := instanceTags[tag]; !exists {
						isSubset = false
						break
					}
				}

				if isSubset {
					clusterMap[key] = struct{}{}
				}
			}
		}
	}

	clusters := make([]Cluster, 0, len(clusterMap))
	for clusterName := range clusterMap {
		cluster := Cluster{
			Name:             clusterName,
			ServiceName:      clusterName,
			Type:             "sds",
			LbType:           "round_robin",
			ConnectTimeoutMs: 1000,
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func buildRoutes(ruleList []rules.Rule) []Route {
	var routes []Route
	for _, rule := range ruleList {
		if rule.Route != nil {
			var headers []Header
			if rule.Match != nil {
				for k, v := range rule.Match.Headers {
					headers = append(
						headers,
						Header{
							Name:  k,
							Value: v,
						},
					)
				}
			}

			for _, backend := range rule.Route.Backends {
				clusterName := BuildServiceName(backend.Name, backend.Tags)

				runtime := &Runtime{
					Key:     backend.Name + "." + BuildServiceName("_", backend.Tags),
					Default: 0,
				}

				route := Route{
					Runtime:       runtime,
					Prefix:        "/" + backend.Name + "/",
					PrefixRewrite: "/",
					Cluster:       clusterName,
					Headers:       headers,
				}

				routes = append(routes, route)
			}
		}
	}

	return routes
}

type ByPriority []rules.Rule

func (s ByPriority) Len() int {
	return len(s)
}

func (s ByPriority) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByPriority) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

func sanitizeRules(ruleList []rules.Rule) {
	for i := range ruleList {
		rule := &ruleList[i]
		if rule.Route != nil {
			var sum float64

			undefined := 0
			for j := range rule.Route.Backends {
				backend := &ruleList[i].Route.Backends[j]
				if backend.Name == "" {
					backend.Name = rule.Destination
				}

				if backend.Weight == 0.0 {
					undefined++
				} else {
					sum += backend.Weight
				}

				sort.Strings(backend.Tags)
			}

			if undefined > 0 {
				w := (1.0 - sum) / float64(undefined)
				for j := range rule.Route.Backends {
					backend := &ruleList[i].Route.Backends[j]
					if backend.Weight == 0 {
						backend.Weight = w
					}
				}
			}
		}
	}

	sort.Sort(sort.Reverse(ByPriority(ruleList))) // Descending order
}

func addDefaultRouteRules(ruleList []rules.Rule, instances []api.ServiceInstance) []rules.Rule {
	serviceMap := make(map[string]struct{})
	for _, instance := range instances {
		serviceMap[instance.ServiceName] = struct{}{}
	}

	for _, rule := range ruleList {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				delete(serviceMap, backend.Name)
			}
		}
	}

	// Provide defaults for all services without any routing rules.
	defaults := make([]rules.Rule, 0, len(serviceMap))
	for service := range serviceMap {
		defaults = append(defaults, rules.Rule{
			Route: &rules.Route{
				Backends: []rules.Backend{
					{
						Name:   service,
						Weight: 1.0,
					},
				},
			},
		})
	}

	return append(ruleList, defaults...)
}

const (
	RuntimePath         = "/etc/envoy/runtime/routing"
	RuntimeVersionsPath = "/etc/envoy/routing_versions"
)

func buildFS(ruleList []rules.Rule) error {
	type weightSpec struct {
		Service string
		Cluster string
		Weight  int
	}

	var weights []weightSpec
	for _, rule := range ruleList {
		if rule.Route != nil {
			w := 0
			for _, backend := range rule.Route.Backends {
				w += int(100 * backend.Weight)
				weight := weightSpec{
					Service: backend.Name,
					Cluster: BuildServiceName("_", backend.Tags),
					Weight:  w,
				}
				weights = append(weights, weight)
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(RuntimePath), 0775); err != nil { // FIXME: hack
		return err
	}

	if err := os.MkdirAll(RuntimeVersionsPath, 0775); err != nil {
		return err
	}

	dirName, err := ioutil.TempDir(RuntimeVersionsPath, "")
	if err != nil {
		return err
	}

	for _, weight := range weights {
		if err := os.MkdirAll(filepath.Join(dirName, "/traffic_shift/", weight.Service), 0775); err != nil {
			return err
		} // FIXME: filemode?

		filename := filepath.Join(dirName, "/traffic_shift/", weight.Service, weight.Cluster)
		data := []byte(fmt.Sprintf("%v", weight.Weight))
		if err := ioutil.WriteFile(filename, data, 0664); err != nil {
			return err
		}
	}

	// FIXME: conflicts
	data := make([]byte, 16)
	for i := range data {
		data[i] = '0' + byte(rand.Intn(10))
	}

	tmpName := "./" + string(data)

	oldRuntime, err := os.Readlink(RuntimePath)
	if err != nil && !os.IsNotExist(err) { // If the error is that the symlink doesn't exist, we ignore it.
		return err
	}

	if err := os.Symlink(dirName, tmpName); err != nil {
		return err
	}

	// Atomically replace the runtime symlink
	if err := os.Rename(tmpName, RuntimePath); err != nil {
		return err
	}

	// Clean up the old config FS if necessary
	// TODO: make this safer
	if oldRuntime != "" {
		oldRuntimeDir := filepath.Dir(oldRuntime)
		if filepath.Clean(oldRuntimeDir) == filepath.Clean(RuntimeVersionsPath) {
			toDelete := filepath.Join(RuntimeVersionsPath, filepath.Base(oldRuntime))
			if err := os.RemoveAll(toDelete); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildFaults(ctlrRules []rules.Rule, serviceName string) ([]HTTPFilter, error) {

	filters := []HTTPFilter{}

	for _, rule := range ctlrRules {
		var headers []HTTPHeader
		if rule.Match != nil {
			headers = make([]HTTPHeader, 0, len(rule.Match.Headers))
			for key, val := range rule.Match.Headers {
				headers = append(headers, HTTPHeader{
					Name:  key,
					Value: val,
				})
			}
		}

		if rule.Destination == serviceName {
			for _, action := range rule.Actions {
				switch action.GetType() {
				case "delay":
					delay := action.Internal().(rules.DelayAction)
					filter := HTTPFilter{
						Type: "decoder",
						Name: "fault",
						Config: &HTTPFilterFaultConfig{
							Delay: &HTTPDelayFilter{
								Type:     "fixed",
								Percent:  int(delay.Probability * 100),
								Duration: delay.Duration,
							},
							Headers: headers,
						},
					}
					filters = append(filters, filter)
				case "abort":
					abort := action.Internal().(rules.AbortAction)
					filter := HTTPFilter{
						Type: "decoder",
						Name: "fault",
						Config: &HTTPFilterFaultConfig{
							Abort: &HTTPAbortFilter{
								Percent:    int(abort.Probability * 100),
								HTTPStatus: abort.ReturnCode,
							},
							Headers: headers,
						},
					}
					filters = append(filters, filter)
				}
			}
		}
	}

	filters = append(filters, HTTPFilter{
		Type:   "decoder",
		Name:   "router",
		Config: HTTPFilterRouterConfig{},
	})
	return filters, nil
}
