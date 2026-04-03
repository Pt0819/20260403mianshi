package output

import (
	"strings"
	"testing"

	"github.com/huashunxinan/mdns-scanner/pkg/models"
)

func TestYAMLFormatter_Format_Empty(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := formatter.Format(nil)
	if result != "services: {}\n" {
		t.Errorf("空结果期望返回'services: {}\\n'，实际得到'%s'", result)
	}

	result = formatter.Format(&models.ScanResult{Services: []models.Service{}})
	if result != "services: {}\n" {
		t.Errorf("空服务列表期望返回'services: {}\\n'，实际得到'%s'", result)
	}
}

func TestYAMLFormatter_Format_SingleService(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "test-device",
				Type:     "_http._tcp.local",
				Port:     80,
				IPv4:     "192.168.1.1",
				IPv6:     "fe80::1",
				Hostname: "test.local",
				TTL:      10,
			},
		},
	}

	output := formatter.Format(result)

	// 验证输出包含必需字段
	if !strings.Contains(output, "80/tcp http") {
		t.Error("输出应包含'80/tcp http'")
	}
	if !strings.Contains(output, "Name=test-device") {
		t.Error("输出应包含Name字段")
	}
	if !strings.Contains(output, "IPv4=192.168.1.1") {
		t.Error("输出应包含IPv4字段")
	}
	if !strings.Contains(output, "IPv6=fe80::1") {
		t.Error("输出应包含IPv6字段")
	}
	if !strings.Contains(output, "Hostname=test.local") {
		t.Error("输出应包含Hostname字段")
	}
	if !strings.Contains(output, "TTL=10") {
		t.Error("输出应包含TTL字段")
	}
}

func TestYAMLFormatter_Format_MultipleServices(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "device1",
				Type:     "_http._tcp.local",
				Port:     80,
				IPv4:     "192.168.1.1",
				Hostname: "device1.local",
				TTL:      10,
			},
			{
				Name:     "device2",
				Type:     "_smb._tcp.local",
				Port:     445,
				IPv4:     "192.168.1.2",
				Hostname: "device2.local",
				TTL:      20,
			},
		},
	}

	output := formatter.Format(result)

	// 验证两个服务都在输出中
	if !strings.Contains(output, "80/tcp http") {
		t.Error("输出应包含HTTP服务")
	}
	if !strings.Contains(output, "445/tcp smb") {
		t.Error("输出应包含SMB服务")
	}
}

func TestYAMLFormatter_Format_WithBanner(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "qnap-nas",
				Type:     "_qdiscover._tcp.local",
				Port:     5000,
				IPv4:     "192.168.1.100",
				Hostname: "qnap.local",
				TTL:      10,
				Banner: map[string]string{
					"model":        "TS-464",
					"fwVer":        "5.2.9",
					"accessType":   "https",
					"accessPort":   "86",
					"displayModel": "TS-464C",
				},
			},
		},
	}

	output := formatter.Format(result)

	// 验证banner信息在输出中
	if !strings.Contains(output, "model=TS-464") {
		t.Error("输出应包含model banner")
	}
	if !strings.Contains(output, "fwVer=5.2.9") {
		t.Error("输出应包含fwVer banner")
	}
	if !strings.Contains(output, "accessType=https") {
		t.Error("输出应包含accessType banner")
	}
}

func TestYAMLFormatter_Format_QNAPExample(t *testing.T) {
	formatter := NewYAMLFormatter()

	// 使用示例中的QNAP设备数据
	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "slw-nas [24:5e:be:69:a3:13]",
				Type:     "_workstation._tcp.local",
				Port:     9,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
			},
			{
				Name:     "slw-nas",
				Type:     "_http._tcp.local",
				Port:     5000,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner:   map[string]string{"path": "/"},
			},
			{
				Name:     "slw-nas",
				Type:     "_qdiscover._tcp.local",
				Port:     5000,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner: map[string]string{
					"accessType":   "https",
					"accessPort":   "86",
					"model":        "TS-X64",
					"displayModel": "TS-464C",
					"fwVer":        "5.2.9",
					"fwBuildNum":   "20260214",
				},
			},
		},
	}

	output := formatter.Format(result)

	// 验证输出格式符合示例
	if !strings.Contains(output, "9/tcp workstation") {
		t.Error("输出应包含workstation服务")
	}
	if !strings.Contains(output, "5000/tcp http") {
		t.Error("输出应包含http服务")
	}
	if !strings.Contains(output, "5000/tcp qdiscover") {
		t.Error("输出应包含qdiscover服务")
	}
	if !strings.Contains(output, "path=/") {
		t.Error("输出应包含path=/")
	}
	if !strings.Contains(output, "accessType=https") {
		t.Error("输出应包含accessType=https")
	}
	if !strings.Contains(output, "model=TS-X64") {
		t.Error("输出应包含model=TS-X64")
	}
	if !strings.Contains(output, "fwVer=5.2.9") {
		t.Error("输出应包含fwVer=5.2.9")
	}

	// 验证PTR记录部分
	if !strings.Contains(output, "answers:") {
		t.Error("输出应包含answers部分")
	}
	if !strings.Contains(output, "PTR:") {
		t.Error("输出应包含PTR部分")
	}
}

func TestYAMLFormatter_Format_PortZero(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "device-info",
				Type:     "_device-info._tcp.local",
				Port:     0,
				IPv4:     "192.168.1.1",
				Hostname: "device.local",
				TTL:      10,
			},
		},
	}

	output := formatter.Format(result)

	// 端口为0时不显示端口号
	if strings.Contains(output, "0/tcp") {
		t.Error("端口为0时不应显示'0/tcp'")
	}
	if !strings.Contains(output, "device-info:") {
		t.Error("输出应包含服务名称")
	}
}

func TestYAMLFormatter_GetServiceName(t *testing.T) {
	formatter := NewYAMLFormatter()

	tests := []struct {
		input    string
		expected string
	}{
		{"_workstation._tcp.local", "workstation"},
		{"_http._tcp.local", "http"},
		{"_https._tcp.local", "https"},
		{"_smb._tcp.local", "smb"},
		{"_afpovertcp._tcp.local", "afpovertcp"},
		{"_qdiscover._tcp.local", "qdiscover"},
		{"_device-info._tcp.local", "device-info"},
		{"", "unknown"},
		{"unknown-service", "unknown-service"},
	}

	for _, test := range tests {
		result := formatter.getServiceName(test.input)
		if result != test.expected {
			t.Errorf("getServiceName(%s) = %s, 期望 %s", test.input, result, test.expected)
		}
	}
}

func TestYAMLFormatter_Format_SortedByPort(t *testing.T) {
	formatter := NewYAMLFormatter()

	// 故意乱序
	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name: "service3",
				Type: "_http._tcp.local",
				Port: 8080,
			},
			{
				Name: "service1",
				Type: "_http._tcp.local",
				Port: 80,
			},
			{
				Name: "service2",
				Type: "_smb._tcp.local",
				Port: 445,
			},
		},
	}

	output := formatter.Format(result)

	// 验证输出按端口排序
	idx80 := strings.Index(output, "80/tcp")
	idx445 := strings.Index(output, "445/tcp")
	idx8080 := strings.Index(output, "8080/tcp")

	if idx80 == -1 || idx445 == -1 || idx8080 == -1 {
		t.Fatal("输出应包含所有服务")
	}

	if !(idx80 < idx445 && idx445 < idx8080) {
		t.Error("服务应按端口排序")
	}
}

func TestYAMLFormatter_FormatCompact(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name: "test",
				Type: "_http._tcp.local",
				Port: 80,
			},
		},
	}

	output := formatter.FormatCompact(result)

	if !strings.Contains(output, "services:") {
		t.Error("紧凑格式应包含services字段")
	}
}

func TestYAMLFormatter_FormatJSON(t *testing.T) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "test",
				Type:     "_http._tcp.local",
				Port:     80,
				IPv4:     "192.168.1.1",
				Hostname: "test.local",
				TTL:      10,
				Banner:   map[string]string{"key": "value"},
			},
		},
	}

	output := formatter.FormatJSON(result)

	// 验证JSON格式
	if !strings.Contains(output, `"services"`) {
		t.Error("JSON输出应包含services字段")
	}
	if !strings.Contains(output, `"name": "test"`) {
		t.Error("JSON输出应包含name字段")
	}
	if !strings.Contains(output, `"banner"`) {
		t.Error("JSON输出应包含banner字段")
	}
}

// 基准测试
func BenchmarkYAMLFormatter_Format(b *testing.B) {
	formatter := NewYAMLFormatter()

	result := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "test-device",
				Type:     "_http._tcp.local",
				Port:     80,
				IPv4:     "192.168.1.1",
				Hostname: "test.local",
				TTL:      10,
				Banner: map[string]string{
					"model": "TS-464",
					"fwVer": "5.2.9",
				},
			},
		},
	}

	for i := 0; i < b.N; i++ {
		formatter.Format(result)
	}
}
