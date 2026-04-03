package models

// Service 表示一个mDNS服务
type Service struct {
	Name     string            `yaml:"name,omitempty" json:"name,omitempty"`
	Type     string            `yaml:"type,omitempty" json:"type,omitempty"`
	Port     int               `yaml:"port,omitempty" json:"port,omitempty"`
	IPv4     string            `yaml:"ipv4,omitempty" json:"ipv4,omitempty"`
	IPv6     string            `yaml:"ipv6,omitempty" json:"ipv6,omitempty"`
	Hostname string            `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	TTL      uint32            `yaml:"ttl,omitempty" json:"ttl,omitempty"`
	Banner   map[string]string `yaml:"banner,omitempty" json:"banner,omitempty"`
	RawTXT   []string          `yaml:"-" json:"-"`
}

// ServiceRecord 表示mDNS服务记录
type ServiceRecord struct {
	Instance   string // 服务实例名
	Service    string // 服务类型 (如 _http._tcp)
	Domain     string // 域名 (如 local)
	HostName   string // 主机名
	Port       int    // 服务端口
	IPv4       string // IPv4地址
	IPv6       string // IPv6地址
	TXTRecords []string // TXT记录
	TTL        uint32 // TTL
}

// ScanResult 表示扫描结果
type ScanResult struct {
	Services []Service `yaml:"services" json:"services"`
}

// PTRRecord PTR记录信息
type PTRRecord struct {
	Name     string
	Target   string
	TTL      uint32
}

// SRVRecord SRV记录信息
type SRVRecord struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
	TTL      uint32
}

// TXTRecord TXT记录信息
type TXTRecord struct {
	Name   string
	Values map[string]string
	TTL    uint32
}