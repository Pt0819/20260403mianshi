package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/huashunxinan/mdns-scanner/pkg/models"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter YAML格式输出器
type YAMLFormatter struct{}

// NewYAMLFormatter 创建新的YAML格式化器
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format 格式化扫描结果为YAML
func (f *YAMLFormatter) Format(result *models.ScanResult) string {
	if result == nil || len(result.Services) == 0 {
		return "services: {}\n"
	}

	var builder strings.Builder
	builder.WriteString("services:\n")

	// 按服务和端口分组
	groups := f.groupServices(result.Services)

	// 按端口排序
	ports := make([]int, 0, len(groups))
	for port := range groups {
		ports = append(ports, port)
	}
	sort.Ints(ports)

	// 生成YAML输出
	for _, port := range ports {
		services := groups[port]
		for _, svc := range services {
			f.formatService(&builder, port, svc)
		}
	}

	// 添加PTR记录部分
	builder.WriteString("\nanswers:\n")
	builder.WriteString("  PTR:\n")
	for _, port := range ports {
		services := groups[port]
		for _, svc := range services {
			if svc.Type != "" {
				builder.WriteString(fmt.Sprintf("    %s\n", svc.Type))
			}
		}
	}

	return builder.String()
}

// groupServices 按端口分组服务
func (f *YAMLFormatter) groupServices(services []models.Service) map[int][]models.Service {
	groups := make(map[int][]models.Service)
	for _, svc := range services {
		port := svc.Port
		if port == 0 {
			port = 0 // 如果没有端口信息，归类到端口0
		}
		groups[port] = append(groups[port], svc)
	}
	return groups
}

// formatService 格式化单个服务
func (f *YAMLFormatter) formatService(builder *strings.Builder, port int, svc models.Service) {
	serviceName := f.getServiceName(svc.Type)

	// 端口为0时不显示端口信息
	if port == 0 {
		builder.WriteString(fmt.Sprintf("  %s:\n", serviceName))
	} else {
		builder.WriteString(fmt.Sprintf("  %d/tcp %s:\n", port, serviceName))
	}

	// Name字段
	if svc.Name != "" {
		builder.WriteString(fmt.Sprintf("    Name=%s\n", svc.Name))
	}

	// IPv4地址
	if svc.IPv4 != "" {
		builder.WriteString(fmt.Sprintf("    IPv4=%s\n", svc.IPv4))
	}

	// IPv6地址
	if svc.IPv6 != "" {
		builder.WriteString(fmt.Sprintf("    IPv6=%s\n", svc.IPv6))
	}

	// Hostname
	if svc.Hostname != "" {
		builder.WriteString(fmt.Sprintf("    Hostname=%s\n", svc.Hostname))
	}

	// TTL
	if svc.TTL > 0 {
		builder.WriteString(fmt.Sprintf("    TTL=%d\n", svc.TTL))
	}

	// Banner信息（深度解析后的详细信息）
	if len(svc.Banner) > 0 {
		f.formatBanner(builder, svc.Banner, serviceName)
	}
}

// formatBanner 格式化banner信息
func (f *YAMLFormatter) formatBanner(builder *strings.Builder, banner map[string]string, serviceName string) {
	// 特殊字段优先级（QNAP风格和其他设备信息）
	priorityFields := []string{
		"path", "accessType", "accessPort", "model", "displayModel",
		"fwVer", "fwBuildNum",
	}

	for _, field := range priorityFields {
		if value, ok := banner[field]; ok {
			builder.WriteString(fmt.Sprintf("    %s=%s\n", field, value))
		}
	}

	// 其他banner字段（排除已处理和内部字段）
	for key, value := range banner {
		alreadyAdded := false
		for _, field := range priorityFields {
			if key == field {
				alreadyAdded = true
				break
			}
		}
		// 跳过内部字段和已添加的字段
		if !alreadyAdded && key != "raw" && key != "device_type" && key != "vendor" &&
		   key != "product_name" && key != "firmware_version" && key != "major_version" &&
		   key != "minor_version" && key != "patch_version" {
			builder.WriteString(fmt.Sprintf("    %s=%s\n", key, value))
		}
	}
}

// getServiceName 从服务类型提取服务名称
func (f *YAMLFormatter) getServiceName(serviceType string) string {
	if serviceType == "" {
		return "unknown"
	}

	// 服务类型映射
	serviceNames := map[string]string{
		"_workstation._tcp.local":  "workstation",
		"_http._tcp.local":        "http",
		"_https._tcp.local":       "https",
		"_smb._tcp.local":         "smb",
		"_afpovertcp._tcp.local":  "afpovertcp",
		"_ssh._tcp.local":         "ssh",
		"_ftp._tcp.local":         "ftp",
		"_printer._tcp.local":     "printer",
		"_airplay._tcp.local":     "airplay",
		"_raop._tcp.local":        "raop",
		"_googlecast._tcp.local":  "googlecast",
		"_spotify-connect._tcp.local": "spotify-connect",
		"_hap._tcp.local":         "hap",
		"_homekit._tcp.local":     "homekit",
		"_daap._tcp.local":        "daap",
		"_dacp._tcp.local":        "dacp",
		"_eppc._tcp.local":        "eppc",
		"_qdiscover._tcp.local":   "qdiscover",
		"_device-info._tcp.local": "device-info",
		"_nas._tcp.local":         "nas",
	}

	if name, ok := serviceNames[serviceType]; ok {
		return name
	}

	// 从服务类型提取名称
	parts := strings.Split(serviceType, ".")
	if len(parts) > 0 {
		name := strings.TrimPrefix(parts[0], "_")
		return name
	}

	return serviceType
}

// FormatCompact 格式化紧凑的YAML输出
func (f *YAMLFormatter) FormatCompact(result *models.ScanResult) string {
	data := struct {
		Services []models.Service `yaml:"services"`
	}{
		Services: result.Services,
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("error: %v\n", err)
	}

	return string(yamlData)
}

// FormatJSON 格式化为JSON输出
func (f *YAMLFormatter) FormatJSON(result *models.ScanResult) string {
	if result == nil {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{\n")
	builder.WriteString("  \"services\": [\n")

	for i, svc := range result.Services {
		builder.WriteString("    {\n")
		builder.WriteString(fmt.Sprintf("      \"name\": \"%s\",\n", svc.Name))
		builder.WriteString(fmt.Sprintf("      \"type\": \"%s\",\n", svc.Type))
		builder.WriteString(fmt.Sprintf("      \"port\": %d,\n", svc.Port))
		builder.WriteString(fmt.Sprintf("      \"ipv4\": \"%s\",\n", svc.IPv4))
		builder.WriteString(fmt.Sprintf("      \"ipv6\": \"%s\",\n", svc.IPv6))
		builder.WriteString(fmt.Sprintf("      \"hostname\": \"%s\",\n", svc.Hostname))
		builder.WriteString(fmt.Sprintf("      \"ttl\": %d,\n", svc.TTL))

		if len(svc.Banner) > 0 {
			builder.WriteString("      \"banner\": {\n")
			j := 0
			for key, value := range svc.Banner {
				if j > 0 {
					builder.WriteString(",\n")
				}
				builder.WriteString(fmt.Sprintf("        \"%s\": \"%s\"", key, value))
				j++
			}
			builder.WriteString("\n      }\n")
		} else {
			builder.WriteString("      \"banner\": null\n")
		}

		builder.WriteString("    }")
		if i < len(result.Services)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("  ]\n")
	builder.WriteString("}\n")

	return builder.String()
}