package parser

import (
	"errors"
	"net"
	"strings"
)

// IPParser 解析IP网段
type IPParser struct{}

// NewIPParser 创建新的IP解析器
func NewIPParser() *IPParser {
	return &IPParser{}
}

// Parse 解析IP网段，支持CIDR格式或IP范围格式
// 支持格式:
// - CIDR: 192.168.1.0/24
// - 范围: 192.168.1.1-192.168.1.255
func (p *IPParser) Parse(input string) ([]string, error) {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "/") {
		return p.parseCIDR(input)
	} else if strings.Contains(input, "-") {
		return p.parseRange(input)
	} else {
		// 单个IP
		ip := net.ParseIP(input)
		if ip == nil {
			return nil, errors.New("无效的IP地址: " + input)
		}
		return []string{input}, nil
	}
}

// parseCIDR 解析CIDR格式的IP网段
func (p *IPParser) parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.New("无效的CIDR格式: " + cidr)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) == 0 {
		return nil, errors.New("CIDR解析后无可用IP: " + cidr)
	}

	return ips, nil
}

// parseRange 解析IP范围格式
func (p *IPParser) parseRange(input string) ([]string, error) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return nil, errors.New("无效的IP范围格式: " + input)
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	if startIP == nil {
		return nil, errors.New("无效的起始IP: " + parts[0])
	}

	endIP := net.ParseIP(strings.TrimSpace(parts[1]))
	if endIP == nil {
		return nil, errors.New("无效的结束IP: " + parts[1])
	}

	// 转换为4字节或16字节整数进行比较
	startIP = startIP.To4()
	endIP = endIP.To4()

	if startIP == nil || endIP == nil {
		return nil, errors.New("仅支持IPv4地址")
	}

	var ips []string
	for ip := startIP; bytesCompare(ip, endIP) <= 0; incIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) == 0 {
		return nil, errors.New("IP范围解析后无可用IP: " + input)
	}

	return ips, nil
}

// inc IP地址递增
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// incIP IPv4地址递增
func incIP(ip []byte) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// bytesCompare 比较两个字节数组
func bytesCompare(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}