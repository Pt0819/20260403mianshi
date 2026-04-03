package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/huashunxinan/mdns-scanner/pkg/models"
	"github.com/miekg/dns"
)

// Config mDNS扫描器配置
type Config struct {
	Timeout    int
	Workers    int
	Verbose    bool
	Interfaces []net.Interface
}

// Scanner mDNS扫描器
type Scanner struct {
	config     Config
	conn       *net.UDPConn
	client     *dns.Client
	mdnsAddr   *net.UDPAddr
}

// NewScanner 创建新的mDNS扫描器
func NewScanner(config Config) *Scanner {
	return &Scanner{
		config:   config,
		client:   &dns.Client{Net: "udp", Timeout: time.Duration(config.Timeout) * time.Second},
		mdnsAddr: &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353},
	}
}

// Scan 执行mDNS扫描
func (s *Scanner) Scan(ips []string, ports []int) (*models.ScanResult, error) {
	if s.config.Verbose {
		fmt.Println("[mDNS] 开始扫描...")
	}

	// 创建结果存储
	results := &models.ScanResult{
		Services: []models.Service{},
	}
	var mu sync.Mutex

	// 发现mDNS服务
	services, err := s.discoverServices()
	if err != nil {
		return nil, fmt.Errorf("服务发现失败: %v", err)
	}

	if s.config.Verbose {
		fmt.Printf("[mDNS] 发现 %d 个服务\n", len(services))
	}

	// 查询每个服务的详细信息
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.Workers)

	for _, svc := range services {
		wg.Add(1)
		go func(service *models.ServiceRecord) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 查询服务详情
			details, err := s.queryServiceDetails(service)
			if err != nil && s.config.Verbose {
				fmt.Printf("[mDNS] 查询服务详情失败: %v\n", err)
			}

			if details != nil {
				// 解析TXT记录中的banner信息
				banner := s.parseTXTRecords(details.TXTRecords)

				// 创建服务对象
				svc := models.Service{
					Name:     details.Instance,
					Type:     details.Service,
					Port:     details.Port,
					IPv4:     details.IPv4,
					IPv6:     details.IPv6,
					Hostname: details.HostName,
					TTL:      details.TTL,
					Banner:   banner,
					RawTXT:   details.TXTRecords,
				}

				mu.Lock()
				results.Services = append(results.Services, svc)
				mu.Unlock()
			}
		}(svc)
	}

	wg.Wait()

	return results, nil
}

// discoverServices 发现局域网中的mDNS服务
func (s *Scanner) discoverServices() ([]*models.ServiceRecord, error) {
	var services []*models.ServiceRecord

	// 常见的mDNS服务类型
	serviceTypes := []string{
		"_workstation._tcp.local",
		"_http._tcp.local",
		"_https._tcp.local",
		"_smb._tcp.local",
		"_afpovertcp._tcp.local",
		"_ssh._tcp.local",
		"_ftp._tcp.local",
		"_printer._tcp.local",
		"_airplay._tcp.local",
		"_raop._tcp.local",
		"_googlecast._tcp.local",
		"_spotify-connect._tcp.local",
		"_hap._tcp.local",
		"_homekit._tcp.local",
		"_daap._tcp.local",
		"_dacp._tcp.local",
		"_eppc._tcp.local",
		"_net-assistant._udp.local",
		"_sleep-proxy._udp.local",
		"_companion-link._tcp.local",
		"_qdiscover._tcp.local",
		"_device-info._tcp.local",
		"_nas._tcp.local",
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.Workers)

	for _, svcType := range serviceTypes {
		wg.Add(1)
		go func(serviceType string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			records, err := s.queryPTR(serviceType)
			if err != nil {
				return
			}

			mu.Lock()
			for _, record := range records {
				services = append(services, &models.ServiceRecord{
					Service: serviceType,
					Instance: record.Target,
				})
			}
			mu.Unlock()
		}(svcType)
	}

	wg.Wait()

	return services, nil
}

// queryPTR 查询PTR记录
func (s *Scanner) queryPTR(serviceType string) ([]*models.PTRRecord, error) {
	// 创建UDP连接
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 构建mDNS查询
	msg := new(dns.Msg)
	msg.SetQuestion(serviceType, dns.TypePTR)
	msg.RecursionDesired = false
	msg.Question = []dns.Question{
		{Name: serviceType, Qtype: dns.TypePTR, Qclass: dns.ClassINET},
	}

	// 发送组播查询
	mdnsAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return nil, errors.New("无法转换为UDP连接")
	}

	// 发送查询
	buf, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	_, err = udpConn.WriteToUDP(buf, mdnsAddr)
	if err != nil {
		return nil, err
	}

	// 设置读取超时
	udpConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	var records []*models.PTRRecord

	// 读取响应
	for {
		responseBuf := make([]byte, 65535)
		n, _, err := udpConn.ReadFromUDP(responseBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			continue
		}

		// 解析响应
		response := new(dns.Msg)
		err = response.Unpack(responseBuf[:n])
		if err != nil {
			continue
		}

		// 提取PTR记录
		for _, rr := range response.Answer {
			if ptr, ok := rr.(*dns.PTR); ok {
				records = append(records, &models.PTRRecord{
					Name:   ptr.Hdr.Name,
					Target: ptr.Ptr,
					TTL:    ptr.Hdr.Ttl,
				})
			}
		}
	}

	return records, nil
}

// queryServiceDetails 查询服务详细信息
func (s *Scanner) queryServiceDetails(service *models.ServiceRecord) (*models.ServiceRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Timeout)*time.Second)
	defer cancel()

	result := &models.ServiceRecord{
		Instance: service.Instance,
		Service:  service.Service,
	}

	// 查询SRV记录
	srv, err := s.querySRV(ctx, service.Instance)
	if err == nil && srv != nil {
		result.Port = int(srv.Port)
		result.HostName = srv.Target
		result.TTL = srv.TTL
	}

	// 查询A记录获取IPv4
	if result.HostName != "" {
		ipv4, err := s.queryA(ctx, result.HostName)
		if err == nil {
			result.IPv4 = ipv4
		}

		// 查询AAAA记录获取IPv6
		ipv6, err := s.queryAAAA(ctx, result.HostName)
		if err == nil {
			result.IPv6 = ipv6
		}
	}

	// 查询TXT记录
	txt, err := s.queryTXT(ctx, service.Instance)
	if err == nil {
		result.TXTRecords = txt
	}

	return result, nil
}

// querySRV 查询SRV记录
func (s *Scanner) querySRV(ctx context.Context, instance string) (*models.SRVRecord, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := new(dns.Msg)
	msg.SetQuestion(instance, dns.TypeSRV)
	msg.RecursionDesired = false

	buf, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return nil, errors.New("无法转换为UDP连接")
	}

	mdnsAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	_, err = udpConn.WriteToUDP(buf, mdnsAddr)
	if err != nil {
		return nil, err
	}

	udpConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	responseBuf := make([]byte, 65535)
	n, _, err := udpConn.ReadFromUDP(responseBuf)
	if err != nil {
		return nil, err
	}

	response := new(dns.Msg)
	err = response.Unpack(responseBuf[:n])
	if err != nil {
		return nil, err
	}

	for _, rr := range response.Answer {
		if srv, ok := rr.(*dns.SRV); ok {
			return &models.SRVRecord{
				Priority: srv.Priority,
				Weight:   srv.Weight,
				Port:     srv.Port,
				Target:   srv.Target,
				TTL:      srv.Hdr.Ttl,
			}, nil
		}
	}

	return nil, errors.New("未找到SRV记录")
}

// queryA 查询A记录
func (s *Scanner) queryA(ctx context.Context, hostname string) (string, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	msg := new(dns.Msg)
	msg.SetQuestion(hostname, dns.TypeA)
	msg.RecursionDesired = false

	buf, err := msg.Pack()
	if err != nil {
		return "", err
	}

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return "", errors.New("无法转换为UDP连接")
	}

	mdnsAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	_, err = udpConn.WriteToUDP(buf, mdnsAddr)
	if err != nil {
		return "", err
	}

	udpConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	responseBuf := make([]byte, 65535)
	n, _, err := udpConn.ReadFromUDP(responseBuf)
	if err != nil {
		return "", err
	}

	response := new(dns.Msg)
	err = response.Unpack(responseBuf[:n])
	if err != nil {
		return "", err
	}

	for _, rr := range response.Answer {
		if a, ok := rr.(*dns.A); ok {
			return a.A.String(), nil
		}
	}

	return "", errors.New("未找到A记录")
}

// queryAAAA 查询AAAA记录
func (s *Scanner) queryAAAA(ctx context.Context, hostname string) (string, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	msg := new(dns.Msg)
	msg.SetQuestion(hostname, dns.TypeAAAA)
	msg.RecursionDesired = false

	buf, err := msg.Pack()
	if err != nil {
		return "", err
	}

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return "", errors.New("无法转换为UDP连接")
	}

	mdnsAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	_, err = udpConn.WriteToUDP(buf, mdnsAddr)
	if err != nil {
		return "", err
	}

	udpConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	responseBuf := make([]byte, 65535)
	n, _, err := udpConn.ReadFromUDP(responseBuf)
	if err != nil {
		return "", err
	}

	response := new(dns.Msg)
	err = response.Unpack(responseBuf[:n])
	if err != nil {
		return "", err
	}

	for _, rr := range response.Answer {
		if aaaa, ok := rr.(*dns.AAAA); ok {
			return aaaa.AAAA.String(), nil
		}
	}

	return "", errors.New("未找到AAAA记录")
}

// queryTXT 查询TXT记录
func (s *Scanner) queryTXT(ctx context.Context, instance string) ([]string, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := new(dns.Msg)
	msg.SetQuestion(instance, dns.TypeTXT)
	msg.RecursionDesired = false

	buf, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return nil, errors.New("无法转换为UDP连接")
	}

	mdnsAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	_, err = udpConn.WriteToUDP(buf, mdnsAddr)
	if err != nil {
		return nil, err
	}

	udpConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	responseBuf := make([]byte, 65535)
	n, _, err := udpConn.ReadFromUDP(responseBuf)
	if err != nil {
		return nil, err
	}

	response := new(dns.Msg)
	err = response.Unpack(responseBuf[:n])
	if err != nil {
		return nil, err
	}

	var txtRecords []string
	for _, rr := range response.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			txtRecords = append(txtRecords, txt.Txt...)
		}
	}

	return txtRecords, nil
}

// parseTXTRecords 解析TXT记录为banner信息
func (s *Scanner) parseTXTRecords(txtRecords []string) map[string]string {
	banner := make(map[string]string)

	for _, txt := range txtRecords {
		// TXT记录格式通常是 key=value
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			banner[key] = value
		}
	}

	return banner
}

// GetServiceName 从服务类型提取服务名称
func (s *Scanner) GetServiceName(serviceType string) string {
	// 从 _http._tcp.local 提取 http
	parts := strings.Split(serviceType, ".")
	if len(parts) > 0 {
		name := strings.TrimPrefix(parts[0], "_")
		return name
	}
	return serviceType
}