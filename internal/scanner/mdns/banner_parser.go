package mdns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BannerParser banner解析器
type BannerParser struct{}

// NewBannerParser 创建新的banner解析器
func NewBannerParser() *BannerParser {
	return &BannerParser{}
}

// Parse 解析banner信息
func (p *BannerParser) Parse(txtRecords []string) map[string]string {
	banner := make(map[string]string)

	for _, txt := range txtRecords {
		// 基本的key=value解析
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			banner[key] = value
		}
	}

	// 深度解析特定设备信息
	p.parseDeviceSpecificInfo(banner)

	return banner
}

// parseDeviceSpecificInfo 解析特定设备的深度信息
func (p *BannerParser) parseDeviceSpecificInfo(banner map[string]string) {
	// QNAS NAS设备信息解析
	if model, ok := banner["model"]; ok {
		if strings.Contains(strings.ToLower(model), "ts") {
			p.parseQNAPBanner(banner)
		}
	}

	// Apple设备信息解析
	if p.isAppleDevice(banner) {
		p.parseAppleBanner(banner)
	}

	// Synology NAS设备信息解析
	if p.isSynologyDevice(banner) {
		p.parseSynologyBanner(banner)
	}

	// 其他设备信息解析
	p.parseGenericDevice(banner)
}

// parseQNAPBanner 解析QNAP NAS设备banner
func (p *BannerParser) parseQNAPBanner(banner map[string]string) {
	// QNAP设备通常包含以下字段：
	// model, displayModel, fwVer, fwBuildNum, accessType, accessPort

	if _, ok := banner["model"]; ok {
		banner["device_type"] = "QNAP NAS"
		banner["vendor"] = "QNAP Systems Inc."
	}

	if displayModel, ok := banner["displayModel"]; ok {
		banner["product_name"] = displayModel
	}

	if fwVer, ok := banner["fwVer"]; ok {
		banner["firmware_version"] = fwVer
		p.parseFirmwareVersion(fwVer, banner)
	}

	if fwBuildNum, ok := banner["fwBuildNum"]; ok {
		banner["firmware_build"] = fwBuildNum
	}

	if accessType, ok := banner["accessType"]; ok {
		banner["supported_protocols"] = accessType
	}

	if accessPort, ok := banner["accessPort"]; ok {
		banner["management_port"] = accessPort
	}

	// 解析QNAP的path字段（如果有）
	if path, ok := banner["path"]; ok {
		banner["web_path"] = path
	}
}

// parseAppleBanner 解析Apple设备banner
func (p *BannerParser) parseAppleBanner(banner map[string]string) {
	banner["device_type"] = "Apple Device"
	banner["vendor"] = "Apple Inc."

	// 解析Mac地址
	if name, ok := banner["Name"]; ok {
		if mac := p.extractMACAddress(name); mac != "" {
			banner["mac_address"] = mac
		}
	}

	// 解析模型信息
	if model, ok := banner["model"]; ok {
		banner["product_model"] = model
	}
}

// parseSynologyBanner 解析Synology NAS设备banner
func (p *BannerParser) parseSynologyBanner(banner map[string]string) {
	banner["device_type"] = "Synology NAS"
	banner["vendor"] = "Synology Inc."

	// Synology通常包含版本信息
	if version, ok := banner["version"]; ok {
		banner["firmware_version"] = version
	}

	if model, ok := banner["model"]; ok {
		banner["product_model"] = model
	}
}

// parseGenericDevice 解析通用设备信息
func (p *BannerParser) parseGenericDevice(banner map[string]string) {
	// 解析版本信息
	if version, ok := banner["ver"]; ok {
		banner["version"] = version
	}

	// 解析产品信息
	if product, ok := banner["product"]; ok {
		banner["product_name"] = product
	}

	// 解析制造商信息
	if manufacturer, ok := banner["manufacturer"]; ok {
		banner["vendor"] = manufacturer
	}

	// 解析序列号
	if serial, ok := banner["serial"]; ok {
		banner["serial_number"] = serial
	}

	// 解析UUID
	if uuid, ok := banner["uuid"]; ok {
		banner["device_uuid"] = uuid
	}
}

// parseFirmwareVersion 解析固件版本
func (p *BannerParser) parseFirmwareVersion(version string, banner map[string]string) {
	// 版本格式可能是: 5.2.9 或 4.5.x 或类似格式
	parts := strings.Split(version, ".")

	if len(parts) >= 1 {
		banner["major_version"] = parts[0]
	}

	if len(parts) >= 2 {
		banner["minor_version"] = parts[1]
	}

	if len(parts) >= 3 {
		banner["patch_version"] = parts[2]
	}
}

// isAppleDevice 判断是否为Apple设备
func (p *BannerParser) isAppleDevice(banner map[string]string) bool {
	if model, ok := banner["model"]; ok {
		appleModels := []string{"Mac", "iPhone", "iPad", "iPod", "Apple TV", "Xserve", "MacBook", "iMac", "Mac Pro", "Mac mini"}
		for _, m := range appleModels {
			if strings.Contains(model, m) {
				return true
			}
		}
	}

	if name, ok := banner["Name"]; ok {
		if strings.Contains(name, "Mac") || strings.Contains(name, "AirPlay") {
			return true
		}
	}

	return false
}

// isSynologyDevice 判断是否为Synology设备
func (p *BannerParser) isSynologyDevice(banner map[string]string) bool {
	if model, ok := banner["model"]; ok {
		return strings.Contains(strings.ToLower(model), "synology") ||
		       strings.Contains(strings.ToLower(model), "ds") ||
		       strings.Contains(strings.ToLower(model), "rs")
	}

	if vendor, ok := banner["vendor"]; ok {
		return strings.Contains(strings.ToLower(vendor), "synology")
	}

	return false
}

// extractMACAddress 从字符串中提取MAC地址
func (p *BannerParser) extractMACAddress(s string) string {
	// MAC地址格式: xx:xx:xx:xx:xx:xx 或 [xx:xx:xx:xx:xx:xx]
	re := regexp.MustCompile(`\[?([0-9A-Fa-f]{2}[:-][0-9A-Fa-f]{2}[:-][0-9A-Fa-f]{2}[:-][0-9A-Fa-f]{2}[:-][0-9A-Fa-f]{2}[:-][0-9A-Fa-f]{2})\]?`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetServiceFriendlyName 获取服务友好名称
func (p *BannerParser) GetServiceFriendlyName(serviceType string) string {
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
		"_raop._tcp.local":        "airplay",
		"_googlecast._tcp.local":  "googlecast",
		"_spotify-connect._tcp.local": "spotify",
		"_hap._tcp.local":         "homekit",
		"_homekit._tcp.local":     "homekit",
		"_daap._tcp.local":        "itunes",
		"_dacp._tcp.local":        "itunes-remote",
		"_eppc._tcp.local":        "remote-desktop",
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

// FormatBanner 格式化banner信息为可读字符串
func (p *BannerParser) FormatBanner(banner map[string]string) string {
	if len(banner) == 0 {
		return ""
	}

	var lines []string

	// 优先显示重要字段
	priorityFields := []string{
		"device_type", "vendor", "product_name", "product_model",
		"firmware_version", "serial_number", "mac_address",
	}

	for _, field := range priorityFields {
		if value, ok := banner[field]; ok {
			lines = append(lines, fmt.Sprintf("%s=%s", field, value))
		}
	}

	// 添加其他字段
	for key, value := range banner {
		alreadyAdded := false
		for _, field := range priorityFields {
			if key == field {
				alreadyAdded = true
				break
			}
		}
		if !alreadyAdded {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return strings.Join(lines, ",")
}

// ParsePortBanner 解析端口扫描的banner
func (p *BannerParser) ParsePortBanner(banner string, port int) map[string]string {
	result := make(map[string]string)

	if banner == "" {
		return result
	}

	// 解析HTTP响应
	if port == 80 || port == 443 || port == 8080 || port == 5000 {
		p.parseHTTPBanner(banner, result)
	}

	// 解析SSH响应
	if port == 22 {
		p.parseSSHBanner(banner, result)
	}

	// 解析FTP响应
	if port == 21 {
		p.parseFTPBanner(banner, result)
	}

	// 如果没有特殊解析器，存储原始banner
	if len(result) == 0 {
		result["raw"] = banner
		result["protocol"] = "tcp"
		result["port"] = strconv.Itoa(port)
	}

	return result
}

// parseHTTPBanner 解析HTTP banner
func (p *BannerParser) parseHTTPBanner(banner string, result map[string]string) {
	result["protocol"] = "http"

	// 提取状态行
	lines := strings.Split(banner, "\n")
	if len(lines) > 0 {
		result["status_line"] = strings.TrimSpace(lines[0])
	}

	// 提取HTTP头
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[strings.ToLower(key)] = value
		}
	}

	// 特殊处理Server头
	if server, ok := result["server"]; ok {
		result["server_software"] = server
		p.parseServerSoftware(server, result)
	}
}

// parseSSHBanner 解析SSH banner
func (p *BannerParser) parseSSHBanner(banner string, result map[string]string) {
	result["protocol"] = "ssh"

	// SSH banner格式: SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
	parts := strings.Split(banner, "-")
	if len(parts) >= 2 {
		result["ssh_version"] = parts[0] + "-" + parts[1]

		if len(parts) >= 3 {
			result["ssh_software"] = parts[2]

			// 提取软件信息
			softwareParts := strings.Split(parts[2], " ")
			if len(softwareParts) > 0 {
				result["software_name"] = softwareParts[0]
			}

			if len(softwareParts) > 1 {
				result["os_info"] = strings.Join(softwareParts[1:], " ")
			}
		}
	}
}

// parseFTPBanner 解析FTP banner
func (p *BannerParser) parseFTPBanner(banner string, result map[string]string) {
	result["protocol"] = "ftp"

	// FTP banner格式: 220 vsftpd 3.0.3 (secure)
	if strings.HasPrefix(banner, "220") {
		parts := strings.SplitN(banner, " ", 3)
		if len(parts) >= 2 {
			result["response_code"] = parts[1]
		}
		if len(parts) >= 3 {
			result["server_info"] = strings.TrimSpace(parts[2])
		}
	}
}

// parseServerSoftware 解析Server头信息
func (p *BannerParser) parseServerSoftware(server string, result map[string]string) {
	lowerServer := strings.ToLower(server)

	// 识别常见的Web服务器
	servers := map[string]string{
		"apache": "Apache",
		"nginx":  "nginx",
		"iis":    "Microsoft IIS",
		"litespeed": "LiteSpeed",
		"caddy":  "Caddy",
		"tomcat": "Apache Tomcat",
		"jetty":  "Jetty",
	}

	for key, value := range servers {
		if strings.Contains(lowerServer, key) {
			result["web_server"] = value
			break
		}
	}

	// 提取版本号
	re := regexp.MustCompile(`(\d+\.\d+(\.\d+)?)`)
	matches := re.FindStringSubmatch(server)
	if len(matches) > 1 {
		result["server_version"] = matches[1]
	}
}