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

	"github.com/faceair/clash-speedtest/unlock"
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
	UnlockTest     string
	Fast           bool
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

		// 过滤掉无效的节点
		proxiesConfig = filterInvalidProxies(proxiesConfig)

		for i, config := range proxiesConfig {
			proxy, err := adapter.ParseProxy(config)
			if err != nil {
				fmt.Println(fmt.Errorf("proxy %d: %w", i, err))
				continue
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
	ProxyName     string                   `json:"proxy_name"`
	ProxyType     string                   `json:"proxy_type"`
	ProxyConfig   map[string]any           `json:"proxy_config"`
	Latency       time.Duration            `json:"latency"`
	Jitter        time.Duration            `json:"jitter"`
	PacketLoss    float64                  `json:"packet_loss"`
	DownloadSize  float64                  `json:"download_size"`
	DownloadTime  time.Duration            `json:"download_time"`
	DownloadSpeed float64                  `json:"download_speed"`
	UploadSize    float64                  `json:"upload_size"`
	UploadTime    time.Duration            `json:"upload_time"`
	UploadSpeed   float64                  `json:"upload_speed"`
	UnlockResults map[string]*UnlockResult `json:"unlock_results,omitempty"`
	IpInfoResult  IpInfo                   `json:"ip_info,omitempty"`
}

type UnlockResult struct {
	Platform string `json:"platform"`
	Status   string `json:"status"`
	Region   string `json:"region,omitempty"`
	Info     string `json:"info,omitempty"`
}

type IpInfo struct {
	Ip          string `json:"ip"`
	Country     string `json:"country"`
	CountryFlag string `json:"flag"`
	Region      string `json:"region,omitempty"`
	City        string `json:"city,omitempty"`
	RiskInfo    string `json:"risk_info,omitempty"`
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

	// 如果延迟测试完全失败或中国联通性检测失败，直接返回
	// 中国联通性检测失败时，packetLoss会被设置为100%
	if result.PacketLoss >= 50 {
		return result
	}

	// 2-3. 并发进行流媒体解锁测试和IP信息获取
	var wg sync.WaitGroup

	// 初始化解锁结果映射
	if st.config.UnlockTest != "" {
		result.UnlockResults = make(map[string]*UnlockResult)
	}

	// 创建通道用于接收流媒体解锁测试结果
	unlockResultChan := make(chan map[string]*unlock.StreamResult, 1)
	// 创建通道用于接收IP信息获取结果
	ipInfoResultChan := make(chan *unlock.IpInfo, 1)

	// 启动流媒体解锁测试
	if st.config.UnlockTest != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 创建HTTP客户端用于解锁测试
			client := st.createClient(proxy)
			// 获取流媒体测试结果
			streamResults := unlock.GetStreamResults(client, st.config.UnlockTest, 50, false)
			unlockResultChan <- streamResults
		}()
	}

	// 启动IP信息获取
	wg.Add(1)
	go func() {
		defer wg.Done()
		ipInfoResult, err := unlock.GetLocationWithRisk(st.createClient(proxy), false, true)
		if err == nil && ipInfoResult != nil {
			ipInfoResultChan <- ipInfoResult
		} else {
			ipInfoResultChan <- &unlock.IpInfo{}
		}
	}()

	// 等待所有并发任务完成
	wg.Wait()

	// 关闭通道
	if st.config.UnlockTest != "" {
		close(unlockResultChan)
	}
	close(ipInfoResultChan)

	// 处理流媒体解锁测试结果
	if st.config.UnlockTest != "" {
		streamResults := <-unlockResultChan
		for platform, streamResult := range streamResults {
			result.UnlockResults[platform] = &UnlockResult{
				Platform: streamResult.Platform,
				Status:   streamResult.Status,
				Region:   streamResult.Region,
				Info:     streamResult.Info,
			}
		}
	}

	// 处理IP信息获取结果
	ipInfoResult := <-ipInfoResultChan
	result.IpInfoResult = IpInfo{
		Ip:          ipInfoResult.Ip,
		Country:     ipInfoResult.Country,
		CountryFlag: ipInfoResult.CountryFlag,
		Region:      ipInfoResult.Region,
		City:        ipInfoResult.City,
		RiskInfo:    ipInfoResult.RiskInfo,
	}
	// 如果是Fast模式，跳过下载和上传测试
	if st.config.Fast {
		return result
	}

	// 4. 并发进行下载和上传测试
	// var wg sync.WaitGroup
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

	// 5. 汇总结果
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
	var failedPingsMutex sync.Mutex

	latencyResults := make(chan time.Duration, 20)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 随机休眠10-210毫秒
			time.Sleep(time.Duration(rand.Intn(200)+10) * time.Millisecond)

			start := time.Now()
			resp, err := client.Get(fmt.Sprintf("%s/__down?bytes=0", st.config.ServerURL))
			if err != nil {
				failedPingsMutex.Lock()
				failedPings++
				failedPingsMutex.Unlock()
				return
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				latencyResults <- time.Since(start)
			} else {
				failedPingsMutex.Lock()
				failedPings++
				failedPingsMutex.Unlock()
			}
		}()
	}

	// 测试server的中国连通性
	if !checkCNNetwork(proxy, client) {
		// 直接返回表示中国连通性失败的结果
		return &latencyResult{
			packetLoss: 100, // 设置为100%丢包率表示完全不可用
			avgLatency: 0,
			jitter:     0,
		}
	}
	// 等待所有ping测试完成
	wg.Wait()
	// 获取最终的failedPings值用于计算
	finalFailedPings := failedPings
	close(latencyResults)

	latencies := make([]time.Duration, 0, len(latencyResults))
	for latency := range latencyResults {
		latencies = append(latencies, latency)
	}

	return calculateLatencyStats(latencies, finalFailedPings)
}

type downloadResult struct {
	bytes    int64
	duration time.Duration
}

func checkCNNetwork(proxy *CProxy, client *http.Client) bool {
	// 获取服务器地址,如果是域名则解析IP
	server := getString(proxy.Config, "server")
	port := getString(proxy.Config, "port")
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
		return checkCnWall(server, port) && checkCnWallBy204(client)
	}
	return checkCnWallBy204(client)
}
func checkCnWallBy204(client *http.Client) bool {
	url := "https://220.185.180.69/generate_204"
	method := "GET"

	payload := &bytes.Buffer{}

	req, _ := http.NewRequest(method, url, payload)

	req.Header.Set("Host", "connectivitycheck.platform.hicloud.com")
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return res.StatusCode == http.StatusNoContent
}

func checkCnWall(ip string, port string) bool {
	url := "https://api.ycwxgzs.com/ipcheck/index.php"
	method := "POST"
	// fmt.Println()
	// fmt.Println("ip:", ip)
	// fmt.Println("port:", port)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("ip", ip)
	_ = writer.WriteField("port", port)
	_ = writer.Close()

	// 设置6秒超时时间
	client := &http.Client{
		Timeout: 6 * time.Second,
	}
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

	// 尝试解析新的JSON响应格式
	type NetworkCheckResponse struct {
		Ip   string `json:"ip"`
		Port string `json:"port"`
		Tcp  string `json:"tcp"`
		Icmp string `json:"icmp"`
	}

	// 解析JSON响应
	var response NetworkCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}

	// 检查TCP或ICMP是否包含"不可用"字样
	return !strings.Contains(response.Tcp, "不可用")
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
		packetLoss: float64(failedPings) / 20.0 * 100,
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

// filterInvalidProxies 过滤掉无效的节点和重复的节点
// 无效节点：类型为ss且cipher属性为ss的节点
// 重复节点：根据节点配置内容判断，如果所有字段内容都相同则视为重复节点，只保留第一个出现的节点
// 注意：即使字段的顺序不同，只要值相同也会被视为重复节点
func filterInvalidProxies(proxiesConfig []map[string]any) []map[string]any {
	filteredProxiesConfig := make([]map[string]any, 0, len(proxiesConfig))
	// 用于存储已添加节点的配置特征
	addedConfigs := make(map[string]bool)

	for _, config := range proxiesConfig {
		// 检查节点类型是否为ss
		if typeValue, ok := config["type"]; ok && typeValue == "ss" {
			// 检查cipher属性是否为ss
			if cipherValue, ok := config["cipher"]; ok && cipherValue == "ss" {
				// 跳过这个节点
				continue
			}
		}

		// 创建一个规范化的配置表示，确保字段顺序不影响比较结果
		// 1. 创建配置的副本，并从中移除name字段（忽略节点名称）
		configCopy := make(map[string]any)
		for k, v := range config {
			// 跳过name字段，不将其加入比较
			if k != "name" {
				configCopy[k] = v
			}
		}

		// 2. 获取所有键并排序
		keys := make([]string, 0, len(configCopy))
		for k := range configCopy {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 3. 按排序后的键顺序创建一个新的有序映射
		orderedMap := make(map[string]any)
		for _, k := range keys {
			orderedMap[k] = configCopy[k]
		}

		// 4. 将有序映射序列化为JSON字符串
		configBytes, err := json.Marshal(orderedMap)
		if err != nil {
			// 如果无法序列化，仍然添加该节点
			filteredProxiesConfig = append(filteredProxiesConfig, config)
			continue
		}
		configStr := string(configBytes)

		// 检查是否已经添加过相同配置的节点
		if _, exists := addedConfigs[configStr]; !exists {
			// 记录这个配置
			addedConfigs[configStr] = true
			// 将有效且不重复的节点添加到过滤后的列表
			filteredProxiesConfig = append(filteredProxiesConfig, config)
		}
	}
	return filteredProxiesConfig
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
