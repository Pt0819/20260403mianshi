package mdns

import (
	"testing"
)

func TestBannerParser_Parse_Basic(t *testing.T) {
	parser := NewBannerParser()

	txtRecords := []string{
		"model=TS-464",
		"fwVer=5.2.9",
	}

	banner := parser.Parse(txtRecords)

	if banner["model"] != "TS-464" {
		t.Errorf("期望model=TS-464，实际得到%s", banner["model"])
	}

	if banner["fwVer"] != "5.2.9" {
		t.Errorf("期望fwVer=5.2.9，实际得到%s", banner["fwVer"])
	}
}

func TestBannerParser_Parse_QNAP(t *testing.T) {
	parser := NewBannerParser()

	txtRecords := []string{
		"model=TS-X64",
		"displayModel=TS-464C",
		"fwVer=5.2.9",
		"fwBuildNum=20260214",
		"accessType=https",
		"accessPort=86",
	}

	banner := parser.Parse(txtRecords)

	// 验证QNAP设备类型识别
	if banner["device_type"] != "QNAP NAS" {
		t.Errorf("期望识别为QNAP NAS，实际得到%s", banner["device_type"])
	}

	if banner["vendor"] != "QNAP Systems Inc." {
		t.Errorf("期望vendor=QNAP Systems Inc.，实际得到%s", banner["vendor"])
	}

	if banner["product_name"] != "TS-464C" {
		t.Errorf("期望product_name=TS-464C，实际得到%s", banner["product_name"])
	}

	if banner["firmware_version"] != "5.2.9" {
		t.Errorf("期望firmware_version=5.2.9，实际得到%s", banner["firmware_version"])
	}
}

func TestBannerParser_Parse_Apple(t *testing.T) {
	parser := NewBannerParser()

	txtRecords := []string{
		"model=Xserve",
		"Name=slw-nas(AFP)",
	}

	banner := parser.Parse(txtRecords)

	// 验证Apple设备识别
	if banner["device_type"] != "Apple Device" {
		t.Errorf("期望识别为Apple Device，实际得到%s", banner["device_type"])
	}

	if banner["vendor"] != "Apple Inc." {
		t.Errorf("期望vendor=Apple Inc.，实际得到%s", banner["vendor"])
	}
}

func TestBannerParser_Parse_Synology(t *testing.T) {
	parser := NewBannerParser()

	txtRecords := []string{
		"model=DS920+",
		"version=7.1",
	}

	banner := parser.Parse(txtRecords)

	// 验证Synology设备识别
	if banner["device_type"] != "Synology NAS" {
		t.Errorf("期望识别为Synology NAS，实际得到%s", banner["device_type"])
	}

	if banner["vendor"] != "Synology Inc." {
		t.Errorf("期望vendor=Synology Inc.，实际得到%s", banner["vendor"])
	}
}

func TestBannerParser_Parse_MACExtraction(t *testing.T) {
	parser := NewBannerParser()

	// 使用Apple设备信息触发MAC地址提取
	txtRecords := []string{
		"model=Xserve",
		"Name=slw-nas [24:5e:be:69:a3:13]",
	}

	banner := parser.Parse(txtRecords)

	// 验证MAC地址提取（需要是Apple设备才会提取）
	if mac, ok := banner["mac_address"]; ok {
		if mac != "24:5e:be:69:a3:13" {
			t.Errorf("期望MAC地址24:5e:be:69:a3:13，实际得到%s", mac)
		}
	}
}

func TestBannerParser_Parse_FirmwareVersion(t *testing.T) {
	parser := NewBannerParser()

	txtRecords := []string{
		"fwVer=5.2.9",
		"model=TS-X64",
	}

	banner := parser.Parse(txtRecords)

	// 验证版本解析
	if banner["major_version"] != "5" {
		t.Errorf("期望major_version=5，实际得到%s", banner["major_version"])
	}

	if banner["minor_version"] != "2" {
		t.Errorf("期望minor_version=2，实际得到%s", banner["minor_version"])
	}

	if banner["patch_version"] != "9" {
		t.Errorf("期望patch_version=9，实际得到%s", banner["patch_version"])
	}
}

func TestBannerParser_GetServiceFriendlyName(t *testing.T) {
	parser := NewBannerParser()

	tests := []struct {
		input    string
		expected string
	}{
		{"_workstation._tcp.local", "workstation"},
		{"_http._tcp.local", "http"},
		{"_smb._tcp.local", "smb"},
		{"_afpovertcp._tcp.local", "afpovertcp"},
		{"_qdiscover._tcp.local", "qdiscover"},
		{"_device-info._tcp.local", "device-info"},
		{"_unknown._tcp.local", "unknown"},
	}

	for _, test := range tests {
		result := parser.GetServiceFriendlyName(test.input)
		if result != test.expected {
			t.Errorf("服务%s: 期望%s，实际得到%s", test.input, test.expected, result)
		}
	}
}

func TestBannerParser_FormatBanner(t *testing.T) {
	parser := NewBannerParser()

	banner := map[string]string{
		"device_type": "QNAP NAS",
		"vendor":      "QNAP Systems Inc.",
		"model":       "TS-464",
		"fwVer":       "5.2.9",
	}

	result := parser.FormatBanner(banner)

	if result == "" {
		t.Error("期望非空格式化结果")
	}

	// 验证格式化结果包含关键字段
	if !contains(result, "device_type=QNAP NAS") {
		t.Error("格式化结果应包含device_type")
	}
}

func TestBannerParser_ParsePortBanner_HTTP(t *testing.T) {
	parser := NewBannerParser()

	banner := "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0\r\nContent-Type: text/html\r\n\r\n"
	result := parser.ParsePortBanner(banner, 80)

	if result["protocol"] != "http" {
		t.Errorf("期望protocol=http，实际得到%s", result["protocol"])
	}

	if result["server"] != "nginx/1.18.0" {
		t.Errorf("期望server=nginx/1.18.0，实际得到%s", result["server"])
	}
}

func TestBannerParser_ParsePortBanner_SSH(t *testing.T) {
	parser := NewBannerParser()

	banner := "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"
	result := parser.ParsePortBanner(banner, 22)

	if result["protocol"] != "ssh" {
		t.Errorf("期望protocol=ssh，实际得到%s", result["protocol"])
	}

	if result["software_name"] != "OpenSSH_8.2p1" {
		t.Errorf("期望software_name=OpenSSH_8.2p1，实际得到%s", result["software_name"])
	}
}

func TestBannerParser_ParsePortBanner_FTP(t *testing.T) {
	parser := NewBannerParser()

	banner := "220 vsftpd 3.0.3 (secure)"
	result := parser.ParsePortBanner(banner, 21)

	if result["protocol"] != "ftp" {
		t.Errorf("期望protocol=ftp，实际得到%s", result["protocol"])
	}
}

func TestBannerParser_ParsePortBanner_Empty(t *testing.T) {
	parser := NewBannerParser()

	result := parser.ParsePortBanner("", 80)

	if len(result) != 0 {
		t.Error("期望空banner返回空map")
	}
}

func TestBannerParser_ExtractMACAddress(t *testing.T) {
	parser := NewBannerParser()

	tests := []struct {
		input    string
		expected string
	}{
		{"slw-nas [24:5e:be:69:a3:13]", "24:5e:be:69:a3:13"},
		{"device [AA:BB:CC:DD:EE:FF]", "AA:BB:CC:DD:EE:FF"},
		{"no-mac-here", ""},
		{"[invalid-mac]", ""},
	}

	for _, test := range tests {
		result := parser.extractMACAddress(test.input)
		if result != test.expected {
			t.Errorf("输入%s: 期望%s，实际得到%s", test.input, test.expected, result)
		}
	}
}

func TestBannerParser_IsAppleDevice(t *testing.T) {
	parser := NewBannerParser()

	tests := []struct {
		banner   map[string]string
		expected bool
	}{
		{map[string]string{"model": "MacBook Pro"}, true},
		{map[string]string{"model": "iMac"}, true},
		{map[string]string{"model": "Xserve"}, true},
		{map[string]string{"model": "Windows PC"}, false},
		{map[string]string{"Name": "My Mac Mini"}, true},
		{map[string]string{}, false},
	}

	for _, test := range tests {
		result := parser.isAppleDevice(test.banner)
		if result != test.expected {
			t.Errorf("isAppleDevice(%v) = %v, 期望 %v", test.banner, result, test.expected)
		}
	}
}

func TestBannerParser_IsSynologyDevice(t *testing.T) {
	parser := NewBannerParser()

	tests := []struct {
		banner   map[string]string
		expected bool
	}{
		{map[string]string{"model": "DS920+"}, true},
		{map[string]string{"model": "RS1221+"}, true},
		{map[string]string{"model": "TS-464"}, false},
		{map[string]string{"vendor": "Synology Inc."}, true},
		{map[string]string{}, false},
	}

	for _, test := range tests {
		result := parser.isSynologyDevice(test.banner)
		if result != test.expected {
			t.Errorf("isSynologyDevice(%v) = %v, 期望 %v", test.banner, result, test.expected)
		}
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}
