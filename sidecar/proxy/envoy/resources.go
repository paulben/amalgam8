package envoy

// AbortFilter definition
type AbortFilter struct {
	Percent    int `json:"abort_percent,omitempty"`
	HTTPStatus int `json:"http_status,omitempty"`
}

// DelayFilter definition
type DelayFilter struct {
	Type     string `json:"type,omitempty"`
	Percent  int    `json:"fixed_delay_percent,omitempty"`
	Duration int    `json:"fixed_duration_ms,omitempty"`
}

// Header definition
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// FilterFaultConfig definition
type FilterFaultConfig struct {
	Abort   *AbortFilter `json:"abort,omitempty"`
	Delay   *DelayFilter `json:"delay,omitempty"`
	Headers []Header     `json:"headers,omitempty"`
}

// FilterRouterConfig definition
type FilterRouterConfig struct {
	DynamicStats bool `json:"dynamic_stats"`
}

// Filter definition
type Filter struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type Runtime struct {
	Key     string `json:"key"`
	Default int    `json:"default"`
}

type Route struct {
	Runtime       *Runtime `json:"runtime,omitempty"`
	Prefix        string   `json:"prefix"`
	PrefixRewrite string   `json:"prefix_rewrite"`
	Cluster       string   `json:"cluster"`
	Headers       []Header `json:"headers,omitempty"`
}

// VirtualHost definition
type VirtualHost struct {
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
	Routes  []Route  `json:"routes"`
}

// RouteConfig definition
type RouteConfig struct {
	VirtualHosts []VirtualHost `json:"virtual_hosts"`
}

// AccessLog definition.
type AccessLog struct {
	Path   string `json:"path"`
	Format string `json:"format,omitempty"`
	Filter string `json:"filter,omitempty"`
}

// NetworkFilterConfig definition
type NetworkFilterConfig struct {
	CodecType         string      `json:"codec_type"`
	StatPrefix        string      `json:"stat_prefix"`
	GenerateRequestID bool        `json:"generate_request_id"`
	RouteConfig       RouteConfig `json:"route_config"`
	Filters           []Filter    `json:"filters"`
	AccessLog         []AccessLog `json:"access_log"`
}

// NetworkFilter definition
type NetworkFilter struct {
	Type   string              `json:"type"`
	Name   string              `json:"name"`
	Config NetworkFilterConfig `json:"config"`
}

// Listener definition
type Listener struct {
	Port    int             `json:"port"`
	Filters []NetworkFilter `json:"filters"`
}

// Admin definition
type Admin struct {
	AccessLogPath string `json:"access_log_path"`
	Port          int    `json:"port"`
}

// Host definition
type Host struct {
	URL string `json:"url"`
}

// Cluster definition
type Cluster struct {
	Name                     string `json:"name"`
	ServiceName              string `json:"service_name,omitempty"`
	ConnectTimeoutMs         int    `json:"connect_timeout_ms"`
	Type                     string `json:"type"`
	LbType                   string `json:"lb_type"`
	MaxRequestsPerConnection int    `json:"max_requests_per_connection,omitempty"`
	Hosts                    []Host `json:"hosts,omitempty"`
}

type ClustersByName []Cluster

func (s ClustersByName) Len() int {
	return len(s)
}

func (s ClustersByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ClustersByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

type SDS struct {
	Cluster        Cluster `json:"cluster"`
	RefreshDelayMs int     `json:"refresh_delay_ms"`
}

// ClusterManager definition
type ClusterManager struct {
	Clusters []Cluster `json:"clusters"`
	SDS      SDS       `json:"sds"`
}

type RootRuntime struct {
	SymlinkRoot          string `json:"symlink_root"`
	Subdirectory         string `json:"subdirectory"`
	OverrideSubdirectory string `json:"override_subdirectory,omitempty"`
}

// Root definition
type Config struct {
	RootRuntime    RootRuntime    `json:"runtime"`
	Listeners      []Listener     `json:"listeners"`
	Admin          Admin          `json:"admin"`
	ClusterManager ClusterManager `json:"cluster_manager"`
}

// ByName implement sort
type ByName []Cluster

// Len length
func (a ByName) Len() int {
	return len(a)
}

// Swap elements
func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less compare
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

//ByHost implement sort
type ByHost []Host

// Len length
func (a ByHost) Len() int {
	return len(a)
}

// Swap elements
func (a ByHost) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less compare
func (a ByHost) Less(i, j int) bool {
	return a[i].URL < a[j].URL
}
