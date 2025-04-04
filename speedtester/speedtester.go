package speedtester

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/provider"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigPaths    string
	FilterRegex    string
	ServerURL      string
	DownloadSize   int
	UploadSize     int
	Timeout        time.Duration
	Concurrent     int
	TestConcurrent int
}

type SpeedTester struct {
	config *Config
}

func New(config *Config) *SpeedTester {
	if config.Concurrent <= 0 {
		config.Concurrent = 1
	}
	if config.DownloadSize <= 0 {
		config.DownloadSize = 100 * 1024 * 1024
	}
	if config.UploadSize <= 0 {
		config.UploadSize = 10 * 1024 * 1024
	}
	if config.TestConcurrent <= 0 {
		config.TestConcurrent = 2
	}
	return &SpeedTester{
		config: config,
	}
}

type CProxy struct {
	constant.Proxy
	Config map[string]any
}

type RawConfig struct {
	Providers map[string]map[string]any `yaml:"proxy-providers"`
	Proxies   []map[string]any          `yaml:"proxies"`
}

func (st *SpeedTester) LoadProxies() (map[string]*CProxy, error) {
	allProxies := make(map[string]*CProxy)

	for _, configPath := range strings.Split(st.config.ConfigPaths, ",") {
		var body []byte
		var err error
		if strings.HasPrefix(configPath, "http") {
			var resp *http.Response
			resp, err = http.Get(configPath)
			if err != nil {
				log.Warnln("failed to fetch config: %s", err)
				continue
			}
			body, err = io.ReadAll(resp.Body)
		} else {
			body, err = os.ReadFile(configPath)
		}
		if err != nil {
			log.Warnln("failed to read config: %s", err)
			continue
		}

		rawCfg := &RawConfig{
			Proxies: []map[string]any{},
		}
		if err := yaml.Unmarshal(body, rawCfg); err != nil {
			return nil, err
		}

		proxies := make(map[string]*CProxy)
		proxiesConfig := rawCfg.Proxies
		providersConfig := rawCfg.Providers

		for i, config := range proxiesConfig {
			proxy, err := adapter.ParseProxy(config)
			if err != nil {
				return nil, fmt.Errorf("proxy %d: %w", i, err)
			}

			if _, exist := proxies[proxy.Name()]; exist {
				return nil, fmt.Errorf("proxy %s is the duplicate name", proxy.Name())
			}
			proxies[proxy.Name()] = &CProxy{Proxy: proxy, Config: config}
		}
		for name, config := range providersConfig {
			if name == provider.ReservedName {
				return nil, fmt.Errorf("can not defined a provider called `%s`", provider.ReservedName)
			}
			pd, err := provider.ParseProxyProvider(name, config)
			if err != nil {
				return nil, fmt.Errorf("parse proxy provider %s error: %w", name, err)
			}
			if err := pd.Initial(); err != nil {
				return nil, fmt.Errorf("initial proxy provider %s error: %w", pd.Name(), err)
			}
			for _, proxy := range pd.Proxies() {
				proxies[fmt.Sprintf("[%s] %s", name, proxy.Name())] = &CProxy{Proxy: proxy}
			}
		}
		for k, p := range proxies {
			switch p.Type() {
			case constant.Shadowsocks, constant.ShadowsocksR, constant.Snell, constant.Socks5, constant.Http,
				constant.Vmess, constant.Vless, constant.Trojan, constant.Hysteria, constant.Hysteria2,
				constant.WireGuard, constant.Tuic, constant.Ssh:
			default:
				continue
			}
			if _, ok := allProxies[k]; !ok {
				allProxies[k] = p
			}
		}
	}

	filterRegexp := regexp.MustCompile(st.config.FilterRegex)
	filteredProxies := make(map[string]*CProxy)
	for name := range allProxies {
		if filterRegexp.MatchString(name) {
			filteredProxies[name] = allProxies[name]
		}
	}
	return filteredProxies, nil
}

func (st *SpeedTester) TestProxies(proxies map[string]*CProxy, fn func(result *Result)) {
	ch := make(chan *Result, len(proxies))

	// 创建一个信号量来控制并发数
	sem := make(chan struct{}, st.config.TestConcurrent)

	// 启动goroutine进行测试
	for name, proxy := range proxies {
		go func(name string, proxy *CProxy) {
			// 获取信号量
			sem <- struct{}{}
			// 测试完成后释放信号量
			defer func() { <-sem }()

			// 执行测试并将结果发送到通道
			ch <- st.testProxy(name, proxy)
		}(name, proxy)
	}

	// 收集所有结果
	for i := 0; i < len(proxies); i++ {
		result := <-ch
		fn(result)
	}

	// 关闭通道
	close(ch)
}

type testJob struct {
	name  string
	proxy *CProxy
}

type Result struct {
	ProxyName     string         `json:"proxy_name"`
	ProxyType     string         `json:"proxy_type"`
	ProxyConfig   map[string]any `json:"proxy_config"`
	Latency       time.Duration  `json:"latency"`
	Jitter        time.Duration  `json:"jitter"`
	PacketLoss    float64        `json:"packet_loss"`
	DownloadSize  float64        `json:"download_size"`
	DownloadTime  time.Duration  `json:"download_time"`
	DownloadSpeed float64        `json:"download_speed"`
	UploadSize    float64        `json:"upload_size"`
	UploadTime    time.Duration  `json:"upload_time"`
	UploadSpeed   float64        `json:"upload_speed"`
}

func (r *Result) FormatDownloadSpeed() string {
	return formatSpeed(r.DownloadSpeed)
}

func (r *Result) FormatLatency() string {
	if r.Latency == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Latency.Milliseconds())
}

func (r *Result) FormatJitter() string {
	if r.Jitter == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", r.Jitter.Milliseconds())
}

func (r *Result) FormatPacketLoss() string {
	return fmt.Sprintf("%.1f%%", r.PacketLoss)
}

func (r *Result) FormatUploadSpeed() string {
	return formatSpeed(r.UploadSpeed)
}

func formatSpeed(bytesPerSecond float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s", "TB/s"}
	unit := 0
	speed := bytesPerSecond
	for speed >= 1024 && unit < len(units)-1 {
		speed /= 1024
		unit++
	}
	return fmt.Sprintf("%.2f%s", speed, units[unit])
}

func (st *SpeedTester) testProxy(name string, proxy *CProxy) *Result {
	result := &Result{
		ProxyName:   name,
		ProxyType:   proxy.Type().String(),
		ProxyConfig: proxy.Config,
	}

	// 1. 首先进行延迟测试
	latencyResult := st.testLatency(proxy)
	result.Latency = latencyResult.avgLatency
	result.Jitter = latencyResult.jitter
	result.PacketLoss = latencyResult.packetLoss

	// 如果延迟测试完全失败，直接返回
	if result.PacketLoss >= 50 {
		return result
	}

	// 2. 并发进行下载和上传测试
	var wg sync.WaitGroup
	downloadResults := make(chan *downloadResult, st.config.Concurrent)

	// 计算每个并发连接的数据大小
	downloadChunkSize := st.config.DownloadSize / st.config.Concurrent
	uploadChunkSize := st.config.UploadSize / st.config.Concurrent

	// 启动下载测试
	for i := 0; i < st.config.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			downloadResults <- st.testDownload(proxy, downloadChunkSize)
		}()
	}

	uploadResults := make(chan *downloadResult, st.config.Concurrent)

	// 启动上传测试
	for i := 0; i < st.config.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			uploadResults <- st.testUpload(proxy, uploadChunkSize)
		}()
	}
	wg.Wait()

	// 3. 汇总结果
	var totalDownloadBytes, totalUploadBytes int64
	var totalDownloadTime, totalUploadTime time.Duration
	var downloadCount, uploadCount int

	for i := 0; i < st.config.Concurrent; i++ {
		if dr := <-downloadResults; dr != nil {
			totalDownloadBytes += dr.bytes
			totalDownloadTime += dr.duration
			downloadCount++
		}
	}
	close(downloadResults)

	for i := 0; i < st.config.Concurrent; i++ {
		if ur := <-uploadResults; ur != nil {
			totalUploadBytes += ur.bytes
			totalUploadTime += ur.duration
			uploadCount++
		}
	}
	close(uploadResults)

	if downloadCount > 0 {
		result.DownloadSize = float64(totalDownloadBytes)
		result.DownloadTime = totalDownloadTime / time.Duration(downloadCount)
		result.DownloadSpeed = float64(totalDownloadBytes) / result.DownloadTime.Seconds()
	}
	if uploadCount > 0 {
		result.UploadSize = float64(totalUploadBytes)
		result.UploadTime = totalUploadTime / time.Duration(uploadCount)
		result.UploadSpeed = float64(totalUploadBytes) / result.UploadTime.Seconds()
	}

	return result
}

type latencyResult struct {
	avgLatency time.Duration
	jitter     time.Duration
	packetLoss float64
}

func (st *SpeedTester) testLatency(proxy *CProxy) *latencyResult {
	client := st.createClient(proxy)
	failedPings := 0

	latencyResults := make(chan time.Duration, 10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 随机休眠10-100毫秒
			time.Sleep(time.Duration(rand.Intn(91)+10) * time.Millisecond)

			start := time.Now()
			resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=0", st.config.ServerURL))
			if err != nil {
				failedPings++
				return
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				latencyResults <- time.Since(start)
			} else {
				failedPings++
			}
		}()
	}
	fmt.Println(checkCNNetwork(proxy))
	//测试server的中国连通性
	if !checkCNNetwork(proxy) {
		failedPings = 10
	}
	wg.Wait()
	if failedPings > 10 {
		failedPings = 10

	}
	close(latencyResults)

	latencies := make([]time.Duration, 0, len(latencyResults))
	for latency := range latencyResults {
		latencies = append(latencies, latency)
	}

	return calculateLatencyStats(latencies, failedPings)
}

type downloadResult struct {
	bytes    int64
	duration time.Duration
}

func checkCNNetwork(proxy *CProxy) bool {
	// 获取服务器地址,如果是域名则解析IP
	server := getString(proxy.Config, "server")
	if server != "" {
		// 检查是否为域名
		if ips, err := net.LookupIP(server); err == nil {
			// 如果能成功解析IP,则使用第一个IP地址
			for _, ip := range ips {
				if ipv4 := ip.To4(); ipv4 != nil {
					server = ipv4.String()
					break
				}
			}
		}
		return sendNetworkRequest(server)
	}
	return false
}

func sendNetworkRequest(ip string) bool {
	url := "https://www.vps234.com/ipcheck/getdata/"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("ip", ip)
	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest(method, url, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}

	// 定义JSON响应结构
	type NetworkCheckData struct {
		InnerICMP bool `json:"innerICMP"`
		InnerTCP  bool `json:"innerTCP"`
		OutICMP   bool `json:"outICMP"`
		OutTCP    bool `json:"outTCP"`
	}

	type NetworkCheckResponse struct {
		Error bool `json:"error"`
		Data  struct {
			Success bool             `json:"success"`
			Msg     string           `json:"msg"`
			Data    NetworkCheckData `json:"data"`
		} `json:"data"`
	}

	// 解析JSON响应
	var response NetworkCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}

	// 检查所有字段是否都为true
	if !response.Error && response.Data.Success {
		netData := response.Data.Data
		return netData.InnerICMP && netData.OutICMP
	}

	return false
}

func (st *SpeedTester) testDownload(proxy constant.Proxy, size int) *downloadResult {
	client := st.createClient(proxy)
	start := time.Now()

	resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=%d", st.config.ServerURL, size))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	downloadBytes, _ := io.Copy(io.Discard, resp.Body)

	return &downloadResult{
		bytes:    downloadBytes,
		duration: time.Since(start),
	}
}

func (st *SpeedTester) testUpload(proxy constant.Proxy, size int) *downloadResult {
	client := st.createClient(proxy)
	reader := NewZeroReader(size)

	start := time.Now()
	resp, err := client.Post(
		fmt.Sprintf("%s/__up", st.config.ServerURL),
		"application/octet-stream",
		reader,
	)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	return &downloadResult{
		bytes:    reader.WrittenBytes(),
		duration: time.Since(start),
	}
}

func (st *SpeedTester) createClient(proxy constant.Proxy) *http.Client {
	return &http.Client{
		Timeout: st.config.Timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				var u16Port uint16
				if port, err := strconv.ParseUint(port, 10, 16); err == nil {
					u16Port = uint16(port)
				}
				return proxy.DialContext(ctx, &constant.Metadata{
					Host:    host,
					DstPort: u16Port,
				})
			},
		},
	}
}

func calculateLatencyStats(latencies []time.Duration, failedPings int) *latencyResult {
	result := &latencyResult{
		packetLoss: float64(failedPings) / 10.0 * 100,
	}

	if len(latencies) == 0 {
		return result
	}

	// 先对延迟数组进行排序
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// 去掉最高和最低值后再计算平均延迟
	var total time.Duration
	count := len(latencies)
	if count > 2 {
		for _, l := range latencies[1 : count-1] {
			total += l
		}
		result.avgLatency = total / time.Duration(count-2)
	} else {
		for _, l := range latencies {
			total += l
		}
		result.avgLatency = total / time.Duration(count)
	}

	// 计算抖动
	var variance float64
	for _, l := range latencies {
		diff := float64(l - result.avgLatency)
		variance += diff * diff
	}
	variance /= float64(len(latencies))
	result.jitter = time.Duration(math.Sqrt(variance))

	return result
}

func getString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			case int:
				return strconv.Itoa(v)
			case float64:
				return fmt.Sprintf("%.0f", v)
			case bool:
				return strconv.FormatBool(v)
			}
		}
	}
	return ""
}
