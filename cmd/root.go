package cmd

import (
	"fmt"
	"os"

	"github.com/huashunxinan/mdns-scanner/internal/parser"
	"github.com/huashunxinan/mdns-scanner/internal/scanner/mdns"
	"github.com/huashunxinan/mdns-scanner/internal/output"
	"github.com/huashunxinan/mdns-scanner/pkg/models"
	"github.com/spf13/cobra"
)

var (
	ipCIDR      string
	portRange   string
	timeout     int
	workers     int
	verbose     bool
	outputFile  string
	demo        bool
)

var rootCmd = &cobra.Command{
	Use:   "mdnsscan",
	Short: "mDNS资产测绘工具",
	Long: `mDNS资产测绘CLI工具 - 扫描指定IP网段和端口范围内的mDNS服务

示例:
  mdnsscan --cidr 192.168.1.0/24 --ports 1-1000
  mdnsscan --cidr 10.0.0.1-10.0.0.255 --ports 80,443,5000
  mdnsscan --cidr 172.16.0.0/16 --ports 1-65535 --workers 100
  mdnsscan --demo  # 演示输出格式`,
	Run: runScan,
}

func init() {
	rootCmd.Flags().StringVarP(&ipCIDR, "cidr", "c", "", "IP网段 (CIDR格式如192.168.1.0/24 或 范围如192.168.1.1-192.168.1.255)")
	rootCmd.Flags().StringVarP(&portRange, "ports", "p", "", "端口范围 (如 1-1000 或 80,443,5000)")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "扫描超时时间(秒)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 50, "并发工作线程数")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出模式")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径 (默认输出到stdout)")
	rootCmd.Flags().BoolVarP(&demo, "demo", "d", false, "演示输出格式（使用模拟数据）")
}

func runScan(cmd *cobra.Command, args []string) {
	// 演示模式
	if demo {
		runDemo()
		return
	}

	// 检查必要参数
	if ipCIDR == "" || portRange == "" {
		fmt.Fprintln(os.Stderr, "错误: 需要指定 --cidr 和 --ports 参数，或使用 --demo 查看演示")
		os.Exit(1)
	}

	// 解析IP网段
	ipParser := parser.NewIPParser()
	targets, err := ipParser.Parse(ipCIDR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析IP网段失败: %v\n", err)
		os.Exit(1)
	}

	// 解析端口范围
	portParser := parser.NewPortParser()
	ports, err := portParser.Parse(portRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析端口范围失败: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("扫描目标: %d 个IP, %d 个端口\n", len(targets), len(ports))
		fmt.Printf("并发数: %d, 超时: %d秒\n", workers, timeout)
	}

	// 创建mDNS扫描器
	mdnsScanner := mdns.NewScanner(mdns.Config{
		Timeout: timeout,
		Workers: workers,
		Verbose: verbose,
	})

	// 执行扫描
	results, err := mdnsScanner.Scan(targets, ports)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描失败: %v\n", err)
		os.Exit(1)
	}

	// 格式化输出
	formatter := output.NewYAMLFormatter()
	outputStr := formatter.Format(results)

	// 输出结果
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(outputStr), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("结果已保存到: %s\n", outputFile)
	} else {
		fmt.Println(outputStr)
	}
}

// runDemo 运行演示模式，展示输出格式
func runDemo() {
	// 创建模拟的扫描结果，展示深度banner识别能力
	results := &models.ScanResult{
		Services: []models.Service{
			{
				Name:     "slw-nas [24:5e:be:69:a3:13]",
				Type:     "_workstation._tcp.local",
				Port:     9,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner:   map[string]string{},
			},
			{
				Name:     "slw-nas",
				Type:     "_http._tcp.local",
				Port:     5000,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner: map[string]string{
					"path": "/",
				},
			},
			{
				Name:     "slw-nas",
				Type:     "_smb._tcp.local",
				Port:     445,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner:   map[string]string{},
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
					"accessType":    "https",
					"accessPort":    "86",
					"model":         "TS-X64",
					"displayModel":  "TS-464C",
					"fwVer":         "5.2.9",
					"fwBuildNum":    "20260214",
					"device_type":   "QNAP NAS",
					"vendor":        "QNAP Systems Inc.",
					"product_name":  "TS-464C",
					"firmware_version": "5.2.9",
				},
			},
			{
				Name:     "slw-nas(AFP)",
				Type:     "_device-info._tcp.local",
				Port:     0,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner: map[string]string{
					"model": "Xserve",
				},
			},
			{
				Name:     "slw-nas(AFP)",
				Type:     "_afpovertcp._tcp.local",
				Port:     548,
				IPv4:     "192.168.1.100",
				IPv6:     "fe80::265e:beff:fe69:a313",
				Hostname: "slw-nas.local",
				TTL:      10,
				Banner:   map[string]string{},
			},
		},
	}

	// 格式化输出
	formatter := output.NewYAMLFormatter()
	outputStr := formatter.Format(results)
	fmt.Println(outputStr)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
