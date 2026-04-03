package port

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/huashunxinan/mdns-scanner/pkg/models"
)

// Config 端口扫描器配置
type Config struct {
	Timeout  int
	Workers  int
	Verbose  bool
	MaxTries int
}

// Scanner 端口扫描器
type Scanner struct {
	config Config
}

// NewScanner 创建新的端口扫描器
func NewScanner(config Config) *Scanner {
	return &Scanner{
		config: config,
	}
}

// ScanResult 端口扫描结果
type ScanResult struct {
	IP       string
	Port     int
	Open     bool
	Banner   string
	Protocol string
}

// Scan 扫描指定IP的端口
func (s *Scanner) Scan(ips []string, ports []int) ([]ScanResult, error) {
	if s.config.Verbose {
		fmt.Printf("[Port] 开始扫描 %d 个IP的 %d 个端口\n", len(ips), len(ports))
	}

	results := make([]ScanResult, 0)
	var mu sync.Mutex

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.Workers)

	for _, ip := range ips {
		for _, port := range ports {
			wg.Add(1)
			go func(targetIP string, targetPort int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// 检查端口是否开放
				open, protocol := s.checkPort(targetIP, targetPort)
				if !open {
					return
				}

				// 获取服务banner
				banner := s.grabBanner(targetIP, targetPort, protocol)

				mu.Lock()
				results = append(results, ScanResult{
					IP:       targetIP,
					Port:     targetPort,
					Open:     true,
					Banner:   banner,
					Protocol: protocol,
				})
				mu.Unlock()

				if s.config.Verbose {
					fmt.Printf("[Port] 发现开放端口: %s:%d (%s)\n", targetIP, targetPort, protocol)
				}
			}(ip, port)
		}
	}

	wg.Wait()

	if s.config.Verbose {
		fmt.Printf("[Port] 扫描完成，发现 %d 个开放端口\n", len(results))
	}

	return results, nil
}

// checkPort 检查端口是否开放
func (s *Scanner) checkPort(ip string, port int) (bool, string) {
	timeout := time.Duration(s.config.Timeout) * time.Second

	// 先尝试TCP连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err == nil {
		conn.Close()
		return true, "tcp"
	}

	// 尝试UDP（对于某些UDP服务）
	conn, err = net.DialTimeout("udp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err == nil {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		conn.Close()
		return true, "udp"
	}

	return false, ""
}

// grabBanner 获取服务banner
func (s *Scanner) grabBanner(ip string, port int, protocol string) string {
	if protocol != "tcp" {
		return ""
	}

	timeout := time.Duration(s.config.Timeout) * time.Second
	address := fmt.Sprintf("%s:%d", ip, port)

	// 尝试连接服务
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(time.Second))

	// 发送特定探针
	probe := s.getProbe(port)
	if probe != "" {
		_, err = conn.Write([]byte(probe))
		if err != nil {
			return ""
		}
	}

	// 读取响应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}

	banner := string(buf[:n])
	return banner
}

// getProbe 获取针对特定端口的探针数据
func (s *Scanner) getProbe(port int) string {
	probes := map[string]string{
		"21":    "",        // FTP - 不需要探针，banner会自动发送
		"22":    "SSH-2.0-MDNS-Scanner\r\n", // SSH
		"23":    "",        // Telnet
		"25":    "EHLO mdns-scanner.local\r\n", // SMTP
		"80":    "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP
		"443":   "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTPS
		"110":   "",        // POP3
		"143":   "",        // IMAP
		"445":   "",        // SMB
		"5000":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP (QNAP)
		"8080":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP Alt
		"8000":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP Alt
		"8888":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP Alt
		"9000":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTP Alt
		"8443":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTPS Alt
		"9443":  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: MDNS-Scanner\r\n\r\n", // HTTPS Alt
		"161":   "",        // SNMP
		"162":   "",        // SNMP Trap
		"514":   "",        // Syslog
		"3306":  "",        // MySQL
		"5432":  "",        // PostgreSQL
		"6379":  "",        // Redis
		"27017": "",        // MongoDB
		"53":    "",        // DNS
		"67":    "",        // DHCP
		"68":    "",        // DHCP
		"69":    "",        // TFTP
		"123":   "",        // NTP
		"137":   "",        // NetBIOS
		"138":   "",        // NetBIOS
		"139":   "",        // NetBIOS
		"389":   "",        // LDAP
		"636":   "",        // LDAPS
		"902":   "",        // VMware
		"2049":  "",        // NFS
		"3389":  "",        // RDP
		"5900":  "",        // VNC
		"5985":  "",        // WinRM
		"5986":  "",        // WinRM over SSL
		"8009":  "",        // Apache JServ
		"8081":  "",        // HTTP Alt
		"8082":  "",        // HTTP Alt
		"8083":  "",        // HTTP Alt
		"8181":  "",        // HTTP Alt
		"9001":  "",        // HTTP Alt
		"9999":  "",        // HTTP Alt
	}

	probe, ok := probes[fmt.Sprintf("%d", port)]
	if !ok {
		return ""
	}
	return probe
}

// IdentifyService 识别服务类型
func (s *Scanner) IdentifyService(banner string, port int) string {
	if banner == "" {
		return "unknown"
	}

	// 常见服务特征匹配
	signatures := map[string]string{
		"SSH-":              "ssh",
		"220":               "ftp",
		"SMTP":              "smtp",
		"HTTP":              "http",
		"Server:":           "http",
		"SMB":               "smb",
		"Microsoft Windows": "smb",
		"Apache":            "http",
		"nginx":             "http",
		"IIS":               "http",
		"MySQL":             "mysql",
		"PostgreSQL":        "postgresql",
		"Redis":             "redis",
		"MongoDB":           "mongodb",
		"JBoss":             "jboss",
		"Tomcat":            "tomcat",
		"Jetty":             "jetty",
		"QNAP":              "qnap-nas",
		"NAS":               "nas",
	}

	for sig, svc := range signatures {
		if contains(banner, sig) {
			return svc
		}
	}

	return "unknown"
}

// contains 检查字符串是否包含子串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

// MergeResults 合并端口扫描结果到mDNS服务列表
func (s *Scanner) MergeResults(services []models.Service, portResults []ScanResult) []models.Service {
	// 创建端口结果映射
	portMap := make(map[string]map[int]ScanResult)
	for _, pr := range portResults {
		if _, ok := portMap[pr.IP]; !ok {
			portMap[pr.IP] = make(map[int]ScanResult)
		}
		portMap[pr.IP][pr.Port] = pr
	}

	// 合并结果
	for i, svc := range services {
		if svc.IPv4 != "" {
			if ports, ok := portMap[svc.IPv4]; ok {
				if portResult, ok := ports[svc.Port]; ok {
					services[i].Banner = parseBannerToMap(portResult.Banner)
				}
			}
		}
	}

	return services
}

// parseBannerToMap 将banner字符串解析为map
func parseBannerToMap(banner string) map[string]string {
	result := make(map[string]string)

	if banner == "" {
		return result
	}

	// 简单的键值对解析
	lines := splitLines(banner)
	for _, line := range lines {
		parts := splitKeyValue(line)
		if len(parts) == 2 {
			result[trim(parts[0])] = trim(parts[1])
		}
	}

	// 如果整个banner不是键值对格式，将其作为原始内容
	if len(result) == 0 && len(banner) > 0 {
		result["raw"] = banner
	}

	return result
}

// splitLines 分割行
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' || c == '\r' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// splitKeyValue 分割键值对
func splitKeyValue(s string) []string {
	for i, c := range s {
		if c == '=' || c == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// trim 去除空白字符
func trim(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}