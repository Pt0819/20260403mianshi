package parser

import (
	"testing"
)

func TestPortParser_Parse_SinglePort(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("80")
	if err != nil {
		t.Fatalf("解析单个端口失败: %v", err)
	}

	if len(ports) != 1 {
		t.Errorf("期望1个端口，实际得到%d个", len(ports))
	}

	if ports[0] != 80 {
		t.Errorf("期望端口80，实际得到%d", ports[0])
	}
}

func TestPortParser_Parse_PortRange(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("1-5")
	if err != nil {
		t.Fatalf("解析端口范围失败: %v", err)
	}

	if len(ports) != 5 {
		t.Errorf("期望5个端口，实际得到%d个", len(ports))
	}

	expectedPorts := []int{1, 2, 3, 4, 5}
	for i, expected := range expectedPorts {
		if ports[i] != expected {
			t.Errorf("索引%d: 期望%d，实际得到%d", i, expected, ports[i])
		}
	}
}

func TestPortParser_Parse_PortList(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("80,443,8080")
	if err != nil {
		t.Fatalf("解析端口列表失败: %v", err)
	}

	if len(ports) != 3 {
		t.Errorf("期望3个端口，实际得到%d个", len(ports))
	}

	expectedPorts := []int{80, 443, 8080}
	for i, expected := range expectedPorts {
		if ports[i] != expected {
			t.Errorf("索引%d: 期望%d，实际得到%d", i, expected, ports[i])
		}
	}
}

func TestPortParser_Parse_MixedFormat(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("80,443,5000-5002,8080")
	if err != nil {
		t.Fatalf("解析混合格式失败: %v", err)
	}

	if len(ports) != 6 {
		t.Errorf("期望6个端口，实际得到%d个", len(ports))
	}

	// 验证端口已排序
	expectedPorts := []int{80, 443, 5000, 5001, 5002, 8080}
	for i, expected := range expectedPorts {
		if ports[i] != expected {
			t.Errorf("索引%d: 期望%d，实际得到%d", i, expected, ports[i])
		}
	}
}

func TestPortParser_Parse_DuplicatePorts(t *testing.T) {
	parser := NewPortParser()

	// 测试重复端口去重
	ports, err := parser.Parse("80,80,443,443")
	if err != nil {
		t.Fatalf("解析重复端口失败: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("期望2个端口（去重后），实际得到%d个", len(ports))
	}
}

func TestPortParser_Parse_PortSorted(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("443,80,8080")
	if err != nil {
		t.Fatalf("解析端口失败: %v", err)
	}

	// 验证端口已排序
	if ports[0] != 80 || ports[1] != 443 || ports[2] != 8080 {
		t.Errorf("端口未正确排序: %v", ports)
	}
}

func TestPortParser_Parse_LargeRange(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("1-100")
	if err != nil {
		t.Fatalf("解析大范围失败: %v", err)
	}

	if len(ports) != 100 {
		t.Errorf("期望100个端口，实际得到%d个", len(ports))
	}

	// 验证第一个和最后一个
	if ports[0] != 1 {
		t.Errorf("第一个端口应该是1，实际是%d", ports[0])
	}
	if ports[99] != 100 {
		t.Errorf("最后一个端口应该是100，实际是%d", ports[99])
	}
}

func TestPortParser_Parse_InvalidPort(t *testing.T) {
	parser := NewPortParser()

	_, err := parser.Parse("abc")
	if err == nil {
		t.Error("期望解析无效端口时返回错误")
	}
}

func TestPortParser_Parse_PortOutOfRange(t *testing.T) {
	parser := NewPortParser()

	_, err := parser.Parse("99999")
	if err == nil {
		t.Error("期望解析超出范围端口时返回错误")
	}
}

func TestPortParser_Parse_PortZero(t *testing.T) {
	parser := NewPortParser()

	_, err := parser.Parse("0")
	if err == nil {
		t.Error("期望解析端口0时返回错误（端口范围1-65535）")
	}
}

func TestPortParser_Parse_NegativePort(t *testing.T) {
	parser := NewPortParser()

	_, err := parser.Parse("-1")
	if err == nil {
		t.Error("期望解析负数端口时返回错误")
	}
}

func TestPortParser_Parse_InvalidRange(t *testing.T) {
	parser := NewPortParser()

	// 起始端口大于结束端口
	_, err := parser.Parse("100-50")
	if err == nil {
		t.Error("期望解析无效范围时返回错误")
	}
}

func TestPortParser_Parse_EmptyString(t *testing.T) {
	parser := NewPortParser()

	_, err := parser.Parse("")
	if err == nil {
		t.Error("期望解析空字符串时返回错误")
	}
}

func TestPortParser_Parse_WithSpaces(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("  80 , 443  ")
	if err != nil {
		t.Fatalf("解析带空格的端口失败: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("期望2个端口，实际得到%d个", len(ports))
	}
}

func TestPortParser_Parse_MaxPort(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse("65535")
	if err != nil {
		t.Fatalf("解析最大端口失败: %v", err)
	}

	if ports[0] != 65535 {
		t.Errorf("期望端口65535，实际得到%d", ports[0])
	}
}

func TestPortParser_Parse_RangeWithSpaces(t *testing.T) {
	parser := NewPortParser()

	ports, err := parser.Parse(" 1 - 5 ")
	if err != nil {
		t.Fatalf("解析带空格的范围失败: %v", err)
	}

	if len(ports) != 5 {
		t.Errorf("期望5个端口，实际得到%d个", len(ports))
	}
}

// 基准测试
func BenchmarkPortParser_Parse_Range(b *testing.B) {
	parser := NewPortParser()
	for i := 0; i < b.N; i++ {
		parser.Parse("1-1000")
	}
}

func BenchmarkPortParser_Parse_List(b *testing.B) {
	parser := NewPortParser()
	for i := 0; i < b.N; i++ {
		parser.Parse("80,443,8080,8443,5000,5001,5002,9000")
	}
}
