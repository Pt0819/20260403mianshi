package parser

import (
	"testing"
)

func TestIPParser_Parse_SingleIP(t *testing.T) {
	parser := NewIPParser()

	ips, err := parser.Parse("192.168.1.1")
	if err != nil {
		t.Fatalf("解析单个IP失败: %v", err)
	}

	if len(ips) != 1 {
		t.Errorf("期望1个IP，实际得到%d个", len(ips))
	}

	if ips[0] != "192.168.1.1" {
		t.Errorf("期望192.168.1.1，实际得到%s", ips[0])
	}
}

func TestIPParser_Parse_CIDR(t *testing.T) {
	parser := NewIPParser()

	// 测试 /30 网段（4个IP）
	ips, err := parser.Parse("192.168.1.0/30")
	if err != nil {
		t.Fatalf("解析CIDR失败: %v", err)
	}

	if len(ips) != 4 {
		t.Errorf("期望4个IP，实际得到%d个", len(ips))
	}

	expectedIPs := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for i, expected := range expectedIPs {
		if ips[i] != expected {
			t.Errorf("索引%d: 期望%s，实际得到%s", i, expected, ips[i])
		}
	}
}

func TestIPParser_Parse_CIDR_Slash24(t *testing.T) {
	parser := NewIPParser()

	// 测试 /24 网段
	ips, err := parser.Parse("10.0.0.0/24")
	if err != nil {
		t.Fatalf("解析/24 CIDR失败: %v", err)
	}

	if len(ips) != 256 {
		t.Errorf("期望256个IP，实际得到%d个", len(ips))
	}

	// 验证第一个和最后一个IP
	if ips[0] != "10.0.0.0" {
		t.Errorf("第一个IP应该是10.0.0.0，实际是%s", ips[0])
	}
	if ips[255] != "10.0.0.255" {
		t.Errorf("最后一个IP应该是10.0.0.255，实际是%s", ips[255])
	}
}

func TestIPParser_Parse_Range(t *testing.T) {
	parser := NewIPParser()

	ips, err := parser.Parse("192.168.1.1-192.168.1.5")
	if err != nil {
		t.Fatalf("解析IP范围失败: %v", err)
	}

	if len(ips) != 5 {
		t.Errorf("期望5个IP，实际得到%d个", len(ips))
	}

	expectedIPs := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
	for i, expected := range expectedIPs {
		if ips[i] != expected {
			t.Errorf("索引%d: 期望%s，实际得到%s", i, expected, ips[i])
		}
	}
}

func TestIPParser_Parse_InvalidCIDR(t *testing.T) {
	parser := NewIPParser()

	_, err := parser.Parse("invalid-cidr")
	if err == nil {
		t.Error("期望解析无效CIDR时返回错误")
	}
}

func TestIPParser_Parse_InvalidIP(t *testing.T) {
	parser := NewIPParser()

	_, err := parser.Parse("999.999.999.999")
	if err == nil {
		t.Error("期望解析无效IP时返回错误")
	}
}

func TestIPParser_Parse_InvalidRange(t *testing.T) {
	parser := NewIPParser()

	// 起始IP大于结束IP
	_, err := parser.Parse("192.168.1.10-192.168.1.5")
	if err == nil {
		t.Error("期望解析无效范围时返回错误")
	}
}

func TestIPParser_Parse_EmptyString(t *testing.T) {
	parser := NewIPParser()

	_, err := parser.Parse("")
	if err == nil {
		t.Error("期望解析空字符串时返回错误")
	}
}

func TestIPParser_Parse_WithSpaces(t *testing.T) {
	parser := NewIPParser()

	// 带空格的输入
	ips, err := parser.Parse("  192.168.1.1  ")
	if err != nil {
		t.Fatalf("解析带空格的IP失败: %v", err)
	}

	if len(ips) != 1 {
		t.Errorf("期望1个IP，实际得到%d个", len(ips))
	}
}

func TestIPParser_Parse_RangeWithSpaces(t *testing.T) {
	parser := NewIPParser()

	// 带空格的范围
	ips, err := parser.Parse(" 192.168.1.1 - 192.168.1.3 ")
	if err != nil {
		t.Fatalf("解析带空格的范围失败: %v", err)
	}

	if len(ips) != 3 {
		t.Errorf("期望3个IP，实际得到%d个", len(ips))
	}
}

// 基准测试
func BenchmarkIPParser_Parse_CIDR(b *testing.B) {
	parser := NewIPParser()
	for i := 0; i < b.N; i++ {
		parser.Parse("192.168.0.0/16")
	}
}

func BenchmarkIPParser_Parse_Range(b *testing.B) {
	parser := NewIPParser()
	for i := 0; i < b.N; i++ {
		parser.Parse("192.168.0.1-192.168.255.254")
	}
}
