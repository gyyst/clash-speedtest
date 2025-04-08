package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/utils"

	"github.com/faceair/clash-speedtest/proxylink"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/metacubex/mihomo/log"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
)

var (
	configPathsConfig = flag.String("c", "", "config file path, also support http(s) url")
	filterRegexConfig = flag.String("f", ".+", "filter proxies by name, use regexp")
	serverURL         = flag.String("server-url", "https://speed.cloudflare.com", "server url")
	downloadSize      = flag.Int("download-size", 50*1024*1024, "download size for testing proxies")
	uploadSize        = flag.Int("upload-size", 20*1024*1024, "upload size for testing proxies")
	timeout           = flag.Duration("timeout", time.Second*5, "timeout for testing proxies")
	concurrent        = flag.Int("concurrent", 4, "download concurrent size")
	testConcurrent    = flag.Int("test-concurrent", 2, "test proxies concurrent size")
	outputPath        = flag.String("output", "result.txt", "output config file path")
	maxLatency        = flag.Duration("max-latency", 800*time.Millisecond, "filter latency greater than this value")
	minDownloadSpeed  = flag.Float64("min-download-speed", 5, "filter speed less than this value(unit: MB/s)")
	minUploadSpeed    = flag.Float64("min-upload-speed", 0, "filter upload speed less than this value(unit: MB/s)")
	maxPacketLoss     = flag.Float64("max-packet-loss", 0, "filter packet loss greater than this value(unit: %)")
	limit             = flag.Int("limit", 0, "limit the number of proxies in output file, 0 means no limit")
	unlockTest        = flag.String("unlock", "", "test streaming media unlock, support: netflix|chatgpt|disney|youtube|all")
	fastMode          = flag.Bool("fast", false, "only test latency, skip download and upload speed test")
	sortFields        = flag.String("sort", "", "sort proxies by fields, support: latency|jitter|packet_loss|download|upload, multiple fields separated by comma, e.g. download,upload")
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	if *configPathsConfig == "" {
		log.Fatalln("please specify the configuration file")
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:    *configPathsConfig,
		FilterRegex:    *filterRegexConfig,
		ServerURL:      *serverURL,
		DownloadSize:   *downloadSize,
		UploadSize:     *uploadSize,
		Timeout:        *timeout,
		Concurrent:     *concurrent,
		TestConcurrent: *testConcurrent,
		UnlockTest:     *unlockTest,
		Fast:           *fastMode,
	})

	allProxies, err := speedTester.LoadProxies()
	if err != nil {
		log.Fatalln("load proxies failed: %v", err)
	}

	bar := progressbar.Default(int64(len(allProxies)), "测试中...")
	results := make([]*speedtester.Result, 0)
	speedTester.TestProxies(allProxies, func(result *speedtester.Result) {
		bar.Add(1)
		bar.Describe(result.ProxyName)
		results = append(results, result)
	})

	// 根据用户指定的字段或默认规则进行排序
	if *sortFields != "" {
		// 解析用户指定的排序字段
		fields := strings.Split(*sortFields, "|")
		sort.Slice(results, func(i, j int) bool {
			// 依次比较每个字段
			for _, field := range fields {
				field = strings.TrimSpace(field)
				switch field {
				case "latency":
					// 处理延迟为0（N/A）的情况
					if results[i].Latency == 0 && results[j].Latency > 0 {
						// 如果i的延迟为0（N/A），而j有有效延迟，则j应该排在前面
						return false
					}
					if results[i].Latency > 0 && results[j].Latency == 0 {
						// 如果i有有效延迟，而j的延迟为0（N/A），则i应该排在前面
						return true
					}
					// 如果两者都有有效延迟或都为N/A，则按延迟值排序（延迟越低越好）
					if results[i].Latency != results[j].Latency {
						return results[i].Latency < results[j].Latency
					}
				case "jitter":
					// 抖动越低越好，所以是小于号
					if results[i].Jitter != results[j].Jitter {
						return results[i].Jitter < results[j].Jitter
					}
				case "packet_loss":
					// 丢包率越低越好，所以是小于号
					if results[i].PacketLoss != results[j].PacketLoss {
						return results[i].PacketLoss < results[j].PacketLoss
					}
				case "download":
					// 下载速度越高越好，所以是大于号
					if results[i].DownloadSpeed != results[j].DownloadSpeed {
						return results[i].DownloadSpeed > results[j].DownloadSpeed
					}
				case "upload":
					// 上传速度越高越好，所以是大于号
					if results[i].UploadSpeed != results[j].UploadSpeed {
						return results[i].UploadSpeed > results[j].UploadSpeed
					}
				}
			}
			// 如果所有指定字段都相等，则按名称排序
			return results[i].ProxyName < results[j].ProxyName
		})
	} else if *fastMode {
		// 在Fast模式下按延迟排序
		sort.Slice(results, func(i, j int) bool {
			// 处理延迟为0（N/A）的情况
			if results[i].Latency == 0 && results[j].Latency > 0 {
				// 如果i的延迟为0（N/A），而j有有效延迟，则j应该排在前面
				return false
			}
			if results[i].Latency > 0 && results[j].Latency == 0 {
				// 如果i有有效延迟，而j的延迟为0（N/A），则i应该排在前面
				return true
			}
			// 如果两者都有有效延迟或都为N/A，则按延迟值排序（延迟越低越好）
			return results[i].Latency < results[j].Latency
		})
	} else {
		// 默认按下载速度排序
		sort.Slice(results, func(i, j int) bool {
			// 如果下载速度不同，按下载速度排序（下载速度越高越好）
			if results[i].DownloadSpeed != results[j].DownloadSpeed {
				return results[i].DownloadSpeed > results[j].DownloadSpeed
			}

			// 如果下载速度相同，处理延迟为0（N/A）的情况
			if results[i].Latency == 0 && results[j].Latency > 0 {
				// 如果i的延迟为0（N/A），而j有有效延迟，则j应该排在前面
				return false
			}
			if results[i].Latency > 0 && results[j].Latency == 0 {
				// 如果i有有效延迟，而j的延迟为0（N/A），则i应该排在前面
				return true
			}

			// 如果两者都有有效延迟或都为N/A，则按延迟值排序（延迟越低越好）
			if results[i].Latency != results[j].Latency {
				return results[i].Latency < results[j].Latency
			}

			// 如果延迟也相同，按名称排序
			return results[i].ProxyName < results[j].ProxyName
		})
	}

	printResults(results)

	if *outputPath != "" {
		err = saveConfig(results)
		if err != nil {
			log.Fatalln("save config file failed: %v", err)
		}
		fmt.Printf("\nsave config file to: %s\n", *outputPath)
	}
}

func printResults(results []*speedtester.Result) {
	table := tablewriter.NewWriter(os.Stdout)

	// 准备表头
	headers := []string{
		"序号",
		"节点名称",
		"类型",
		"延迟",
		"抖动",
		"丢包率",
		"风险值",
	}

	// 如果不是Fast模式，添加速度相关列
	if !*fastMode {
		headers = append(headers, "下载速度", "上传速度")
	}

	// 检查是否有解锁测试结果，如果有，添加解锁测试结果列
	hasUnlockResults := false
	for _, result := range results {
		if result.UnlockResults != nil && len(result.UnlockResults) > 0 {
			hasUnlockResults = true
			break
		}
	}

	if hasUnlockResults {
		headers = append(headers, "解锁测试")
	}

	table.SetHeader(headers)

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for i, result := range results {
		idStr := fmt.Sprintf("%d.", i+1)

		// 延迟颜色
		latencyStr := result.FormatLatency()
		if result.Latency > 0 {
			if result.Latency < 800*time.Millisecond {
				latencyStr = colorGreen + latencyStr + colorReset
			} else if result.Latency < 1500*time.Millisecond {
				latencyStr = colorYellow + latencyStr + colorReset
			} else {
				latencyStr = colorRed + latencyStr + colorReset
			}
		} else {
			latencyStr = colorRed + latencyStr + colorReset
		}

		jitterStr := result.FormatJitter()
		if result.Jitter > 0 {
			if result.Jitter < 800*time.Millisecond {
				jitterStr = colorGreen + jitterStr + colorReset
			} else if result.Jitter < 1500*time.Millisecond {
				jitterStr = colorYellow + jitterStr + colorReset
			} else {
				jitterStr = colorRed + jitterStr + colorReset
			}
		} else {
			jitterStr = colorRed + jitterStr + colorReset
		}

		// 丢包率颜色
		packetLossStr := result.FormatPacketLoss()
		if result.PacketLoss < 10 {
			packetLossStr = colorGreen + packetLossStr + colorReset
		} else if result.PacketLoss < 20 {
			packetLossStr = colorYellow + packetLossStr + colorReset
		} else {
			packetLossStr = colorRed + packetLossStr + colorReset
		}

		// 下载速度颜色 (以MB/s为单位判断)
		downloadSpeed := result.DownloadSpeed / (1024 * 1024)
		downloadSpeedStr := result.FormatDownloadSpeed()
		if downloadSpeed >= 10 {
			downloadSpeedStr = colorGreen + downloadSpeedStr + colorReset
		} else if downloadSpeed >= 5 {
			downloadSpeedStr = colorYellow + downloadSpeedStr + colorReset
		} else {
			downloadSpeedStr = colorRed + downloadSpeedStr + colorReset
		}

		// 上传速度颜色
		uploadSpeed := result.UploadSpeed / (1024 * 1024)
		uploadSpeedStr := result.FormatUploadSpeed()
		if uploadSpeed >= 5 {
			uploadSpeedStr = colorGreen + uploadSpeedStr + colorReset
		} else if uploadSpeed >= 2 {
			uploadSpeedStr = colorYellow + uploadSpeedStr + colorReset
		} else {
			uploadSpeedStr = colorRed + uploadSpeedStr + colorReset
		}

		// 风险值颜色
		riskInfoStr := "N/A"
		if result.IpInfoResult.RiskInfo != "" {
			riskInfoStr = result.IpInfoResult.RiskInfo
			// 根据风险信息内容设置颜色
			if strings.Contains(riskInfoStr, "较差") || strings.Contains(riskInfoStr, "高危") {
				riskInfoStr = colorRed + riskInfoStr + colorReset
			} else if strings.Contains(riskInfoStr, "一般") || strings.Contains(riskInfoStr, "中危") {
				riskInfoStr = colorYellow + riskInfoStr + colorReset
			} else {
				riskInfoStr = colorGreen + riskInfoStr + colorReset
			}
		} else {
			riskInfoStr = colorGreen + riskInfoStr + colorReset
		}

		row := []string{
			idStr,
			result.ProxyName,
			result.ProxyType,
			latencyStr,
			jitterStr,
			packetLossStr,
			riskInfoStr,
		}

		// 如果不是Fast模式，添加速度相关列
		if !*fastMode {
			row = append(row, downloadSpeedStr, uploadSpeedStr)
		}

		// 如果有解锁测试结果，添加解锁测试结果列
		if hasUnlockResults {
			unlockStr := ""
			if result.UnlockResults != nil && len(result.UnlockResults) > 0 {
				unlockResults := make([]string, 0)
				for platform, unlockResult := range result.UnlockResults {
					if unlockResult.Status == "Success" {
						regionInfo := ""
						if unlockResult.Region != "" {
							regionInfo = "(" + unlockResult.Region + ")"
						}
						unlockResults = append(unlockResults, colorGreen+platform+regionInfo+colorReset)
					} else {
						unlockResults = append(unlockResults, colorRed+platform+colorReset)
					}
				}
				unlockStr = strings.Join(unlockResults, ", ")
			}
			row = append(row, unlockStr)
		}

		table.Append(row)
	}

	fmt.Println()
	table.Render()
	fmt.Println()
}

func saveConfig(results []*speedtester.Result) error {
	filteredResults := make([]*speedtester.Result, 0)
	for _, result := range results {
		if *maxLatency > 0 && result.Latency > *maxLatency {
			continue
		}
		// 在Fast模式下不根据速度过滤
		if !*fastMode {
			if *minDownloadSpeed > 0 && float64(result.DownloadSpeed)/(1024*1024) < *minDownloadSpeed {
				continue
			}
			if *minUploadSpeed > 0 && float64(result.UploadSpeed)/(1024*1024) < *minUploadSpeed {
				continue
			}
		}
		if result.PacketLoss > *maxPacketLoss {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	// 应用limit参数限制代理数量
	if *limit > 0 && len(filteredResults) > *limit {
		filteredResults = filteredResults[:*limit]
	}

	// 创建文本内容，每行一个代理链接
	lines := make([]string, 0, len(filteredResults))
	countryCount := make(map[string]int)
	for _, result := range filteredResults {
		// 构建新的节点名称格式
		originalName := result.ProxyName
		newName := ""

		// 添加国旗和国家信息
		if result.IpInfoResult.Country != "" {
			// 添加国旗
			if result.IpInfoResult.CountryFlag != "" {
				newName += result.IpInfoResult.CountryFlag
			}
			// 添加中文国家名称
			if chineseName, ok := utils.CountryCodeMap[result.IpInfoResult.Country]; ok {
				countryCount[chineseName] += 1
				newName += chineseName
				if countryCount[chineseName] >= 1 {
					newName += fmt.Sprintf("%d", countryCount[chineseName])
				}
			}
			// 添加风险信息
			if result.IpInfoResult.RiskInfo != "" {
				newName += " " + result.IpInfoResult.RiskInfo
			}
			// 添加地区信息
			if result.IpInfoResult.Region != "" && result.IpInfoResult.Region != "N/A" {
				newName += " " + result.IpInfoResult.Region
			}
			if result.IpInfoResult.City != "" && result.IpInfoResult.City != "N/A" {
				newName += " " + result.IpInfoResult.City
			}
		} else {
			// 如果没有国家信息，使用原始名称
			newName = originalName
		}

		// 添加流媒体解锁信息
		if result.UnlockResults != nil && len(result.UnlockResults) > 0 {
			unlockResults := make([]string, 0)
			for platform, unlockResult := range result.UnlockResults {
				if unlockResult.Status == "Success" {
					regionInfo := ""
					if unlockResult.Region != "" {
						regionInfo = "(" + unlockResult.Region + ")"
					}
					unlockResults = append(unlockResults, platform+regionInfo)
				}
			}

			if len(unlockResults) > 0 {
				newName += " [" + strings.Join(unlockResults, "| ") + "]"
			}
		}
		proxyName := newName
		link, err := proxylink.GenerateProxyLink(proxyName, result.ProxyType, result.ProxyConfig)
		if err != nil {
			// 如果生成链接失败，使用代理名称
			link = proxyName
		} else {
			// 对URL进行解码处理
			decodedLink, err := url.QueryUnescape(link)
			if err == nil {
				link = decodedLink
			}
		}
		// 将代理链接添加到文本行中
		lines = append(lines, link)
	}

	// 将所有行合并为一个字符串，每行一个代理链接
	txtData := strings.Join(lines, "\n")

	// 写入文件
	return os.WriteFile(*outputPath, []byte(txtData), 0o644)
}
