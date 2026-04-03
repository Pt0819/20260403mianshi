package parser

import (
	"errors"
	"strconv"
	"strings"
)

// PortParser 解析端口范围
type PortParser struct{}

// NewPortParser 创建新的端口解析器
func NewPortParser() *PortParser {
	return &PortParser{}
}

// Parse 解析端口范围，支持多种格式
// 支持格式:
// - 单个端口: 80
// - 端口范围: 1-1000
// - 端口列表: 80,443,5000
// - 混合格式: 80,443,5000-6000,8080
func (p *PortParser) Parse(input string) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("端口范围不能为空")
	}

	var ports []int
	seen := make(map[int]bool)

	// 按逗号分割
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// 解析端口范围
			rangePorts, err := p.parsePortRange(part)
			if err != nil {
				return nil, err
			}
			for _, port := range rangePorts {
				if !seen[port] {
					seen[port] = true
					ports = append(ports, port)
				}
			}
		} else {
			// 解析单个端口
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, errors.New("无效的端口号: " + part)
			}
			if !p.isValidPort(port) {
				return nil, errors.New("端口号超出范围(1-65535): " + part)
			}
			if !seen[port] {
				seen[port] = true
				ports = append(ports, port)
			}
		}
	}

	// 按端口排序
	p.sortPorts(ports)

	if len(ports) == 0 {
		return nil, errors.New("未找到有效的端口")
	}

	return ports, nil
}

// parsePortRange 解析端口范围格式 "start-end"
func (p *PortParser) parsePortRange(input string) ([]int, error) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return nil, errors.New("无效的端口范围格式: " + input)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, errors.New("无效的起始端口: " + parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, errors.New("无效的结束端口: " + parts[1])
	}

	if !p.isValidPort(start) {
		return nil, errors.New("起始端口超出范围(1-65535)")
	}

	if !p.isValidPort(end) {
		return nil, errors.New("结束端口超出范围(1-65535)")
	}

	if start > end {
		return nil, errors.New("起始端口不能大于结束端口")
	}

	var ports []int
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}

	return ports, nil
}

// isValidPort 检查端口是否有效
func (p *PortParser) isValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

// sortPorts 对端口进行排序
func (p *PortParser) sortPorts(ports []int) {
	// 简单的冒泡排序
	n := len(ports)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if ports[j] > ports[j+1] {
				ports[j], ports[j+1] = ports[j+1], ports[j]
			}
		}
	}
}