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
	sortFields        = flag.String("sort", "weighted", "sort proxies by fields, support: latency|jitter|packet_loss|download|upload|weighted, multiple fields separated by comma, e.g. download,upload")
	renameMode        = flag.String("rename", "overwrite", "rename mode for proxy names: add|overwrite|none")
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

// 计算节点的加权得分
// 返回一个综合得分，得分越低表示性能越好
func calculateWeightedScore(results []*speedtester.Result, index int) float64 {
	// 根据是否为Fast模式定义不同的权重
	var (
		latencyWeight    float64
		jitterWeight     float64
		packetLossWeight float64
		downloadWeight   float64
		uploadWeight     float64
	)

	if *fastMode {
		// Fast模式下只考虑延迟、抖动和丢包率
		latencyWeight = 0.60    // 延迟权重
		jitterWeight = 0.20     // 抖动权重
		packetLossWeight = 0.20 // 丢包率权重
		downloadWeight = 0      // 下载速度权重
		uploadWeight = 0        // 上传速度权重
	} else {
		// 正常模式下考虑所有指标
		latencyWeight = 0.35    // 延迟权重
		jitterWeight = 0.15     // 抖动权重
		packetLossWeight = 0.15 // 丢包率权重
		downloadWeight = 0.30   // 下载速度权重
		uploadWeight = 0.05     // 上传速度权重
	}

	// 创建各项指标的得分映射
	latencyScores := make(map[int]float64)
	jitterScores := make(map[int]float64)
	packetLossScores := make(map[int]float64)
	downloadScores := make(map[int]float64)
	uploadScores := make(map[int]float64)

	// 计算延迟得分（值越小得分越高）
	// 首先找出有效的最大和最小延迟值
	minLatency, maxLatency := float64(0), float64(0)
	hasValidLatency := false
	for _, r := range results {
		if r.Latency > 0 { // 只考虑有效延迟
			if !hasValidLatency {
				minLatency = float64(r.Latency)
				maxLatency = float64(r.Latency)
				hasValidLatency = true
			} else {
				if float64(r.Latency) < minLatency {
					minLatency = float64(r.Latency)
				}
				if float64(r.Latency) > maxLatency {
					maxLatency = float64(r.Latency)
				}
			}
		}
	}

	// 计算延迟得分
	latencyRange := maxLatency - minLatency
	for i, r := range results {
		if r.Latency > 0 && hasValidLatency && latencyRange > 0 {
			// 归一化得分：0是最差，1是最好
			latencyScores[i] = 1.0 - (float64(r.Latency)-minLatency)/latencyRange
		} else {
			// 无效延迟给予最低分
			latencyScores[i] = 0.0
		}
	}

	// 计算抖动得分（值越小得分越高）
	minJitter, maxJitter := float64(0), float64(0)
	hasValidJitter := false
	for _, r := range results {
		if r.Jitter > 0 { // 只考虑有效抖动
			if !hasValidJitter {
				minJitter = float64(r.Jitter)
				maxJitter = float64(r.Jitter)
				hasValidJitter = true
			} else {
				if float64(r.Jitter) < minJitter {
					minJitter = float64(r.Jitter)
				}
				if float64(r.Jitter) > maxJitter {
					maxJitter = float64(r.Jitter)
				}
			}
		}
	}

	// 计算抖动得分
	jitterRange := maxJitter - minJitter
	for i, r := range results {
		if r.Jitter > 0 && hasValidJitter && jitterRange > 0 {
			// 归一化得分：0是最差，1是最好
			jitterScores[i] = 1.0 - (float64(r.Jitter)-minJitter)/jitterRange
		} else {
			// 无效抖动给予最低分
			jitterScores[i] = 0.0
		}
	}

	// 计算丢包率得分（值越小得分越高）
	minPacketLoss, maxPacketLoss := float64(0), float64(0)
	hasValidPacketLoss := false
	for _, r := range results {
		if !hasValidPacketLoss {
			minPacketLoss = r.PacketLoss
			maxPacketLoss = r.PacketLoss
			hasValidPacketLoss = true
		} else {
			if r.PacketLoss < minPacketLoss {
				minPacketLoss = r.PacketLoss
			}
			if r.PacketLoss > maxPacketLoss {
				maxPacketLoss = r.PacketLoss
			}
		}
	}

	// 计算丢包率得分
	packetLossRange := maxPacketLoss - minPacketLoss
	for i, r := range results {
		if hasValidPacketLoss && packetLossRange > 0 {
			// 归一化得分：0是最差，1是最好
			packetLossScores[i] = 1.0 - (r.PacketLoss-minPacketLoss)/packetLossRange
		} else {
			// 如果所有节点丢包率相同，则都给满分
			packetLossScores[i] = 1.0
		}
	}

	// 计算下载速度得分（值越大得分越高）
	minDownload, maxDownload := float64(0), float64(0)
	hasValidDownload := false
	for _, r := range results {
		if !hasValidDownload {
			minDownload = r.DownloadSpeed
			maxDownload = r.DownloadSpeed
			hasValidDownload = true
		} else {
			if r.DownloadSpeed < minDownload {
				minDownload = r.DownloadSpeed
			}
			if r.DownloadSpeed > maxDownload {
				maxDownload = r.DownloadSpeed
			}
		}
	}

	// 计算下载速度得分
	downloadRange := maxDownload - minDownload
	for i, r := range results {
		if hasValidDownload && downloadRange > 0 {
			// 归一化得分：0是最差，1是最好
			downloadScores[i] = (r.DownloadSpeed - minDownload) / downloadRange
		} else {
			// 如果所有节点下载速度相同，则都给满分
			downloadScores[i] = 1.0
		}
	}

	// 计算上传速度得分（值越大得分越高）
	minUpload, maxUpload := float64(0), float64(0)
	hasValidUpload := false
	for _, r := range results {
		if !hasValidUpload {
			minUpload = r.UploadSpeed
			maxUpload = r.UploadSpeed
			hasValidUpload = true
		} else {
			if r.UploadSpeed < minUpload {
				minUpload = r.UploadSpeed
			}
			if r.UploadSpeed > maxUpload {
				maxUpload = r.UploadSpeed
			}
		}
	}

	// 计算上传速度得分
	uploadRange := maxUpload - minUpload
	for i, r := range results {
		if hasValidUpload && uploadRange > 0 {
			// 归一化得分：0是最差，1是最好
			uploadScores[i] = (r.UploadSpeed - minUpload) / uploadRange
		} else {
			// 如果所有节点上传速度相同，则都给满分
			uploadScores[i] = 1.0
		}
	}

	// 计算加权总分（得分越高越好）
	totalScore := latencyScores[index]*latencyWeight +
		jitterScores[index]*jitterWeight +
		packetLossScores[index]*packetLossWeight +
		downloadScores[index]*downloadWeight +
		uploadScores[index]*uploadWeight

	// 返回负值，因为排序时得分越低越好
	return -totalScore
}

func main() {
	flag.Parse()
	log.SetLevel(log.SILENT)

	if *configPathsConfig == "" {
		log.Fatalln("please specify the configuration file")
	}

	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      *configPathsConfig,
		FilterRegex:      *filterRegexConfig,
		ServerURL:        *serverURL,
		DownloadSize:     *downloadSize,
		UploadSize:       *uploadSize,
		Timeout:          *timeout,
		Concurrent:       *concurrent,
		TestConcurrent:   *testConcurrent,
		UnlockTest:       *unlockTest,
		Fast:             *fastMode,
		MaxLatency:       *maxLatency,
		MinDownloadSpeed: *minDownloadSpeed,
		MinUploadSpeed:   *minUploadSpeed,
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
				case "weighted":
					// 加权排序，综合考虑各项指标
					// 计算节点i的加权得分
					scoreI := calculateWeightedScore(results, i)
					// 计算节点j的加权得分
					scoreJ := calculateWeightedScore(results, j)
					// 得分越低越好（注意：calculateWeightedScore返回的是负值，所以这里用<比较）
					if scoreI != scoreJ {
						return scoreI < scoreJ
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

		// 根据rename模式处理节点名称
		switch *renameMode {
		case "none":
			// 不重命名，使用原始名称
			newName = originalName
		case "add":
			// 在原始名称后添加新信息
			newName = originalName
			// 添加国旗和国家信息
			if result.IpInfoResult.Country != "" {
				// 添加国旗
				if result.IpInfoResult.CountryFlag != "" {
					newName += " " + result.IpInfoResult.CountryFlag
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
			}
			// 添加下载和上传速度信息（非Fast模式下）
			if !*fastMode {
				// 格式化下载和上传速度
				downloadSpeedStr := result.FormatDownloadSpeed()
				uploadSpeedStr := result.FormatUploadSpeed()
				// 添加到节点名称中
				newName += fmt.Sprintf(" ⬇%s ⬆%s", downloadSpeedStr, uploadSpeedStr)
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
		case "overwrite":
			fallthrough
		default:
			// 默认为overwrite模式，完全重写节点名称
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
			// 添加下载和上传速度信息（非Fast模式下）
			if !*fastMode {
				// 格式化下载和上传速度
				downloadSpeedStr := result.FormatDownloadSpeed()
				uploadSpeedStr := result.FormatUploadSpeed()
				// 添加到节点名称中
				newName += fmt.Sprintf(" ⬇%s ⬆%s", downloadSpeedStr, uploadSpeedStr)
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
		}

		proxyName := newName
		// 更新代理配置中的名称
		// result.ProxyConfig["name"] = proxyName

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

	// // 尝试使用ParseProxiesJSON方法批量生成URLs
	// // 将filteredResults转换为JSON格式
	// proxiesJSON := make([]map[string]any, 0, len(filteredResults))
	// for _, result := range filteredResults {
	// 	// 确保每个代理配置都有name字段
	// 	proxyCfg := result.ProxyConfig
	// 	proxiesJSON = append(proxiesJSON, proxyCfg)
	// }

	// // 将代理配置数组转换为JSON字节数组
	// jsonData, err := json.Marshal(proxiesJSON)
	// if err == nil {
	// 	// 使用ParseProxiesJSON生成URL格式
	// 	buffer, err := proxylink.ParseProxiesJSON(jsonData)
	// 	if err == nil && buffer.Len() > 0 {
	// 		// 如果成功生成URLs，使用生成的结果替代之前的lines
	// 		txtData := buffer.String()
	// 		return os.WriteFile(*outputPath, []byte(txtData), 0o644)
	// 	}
	// }

	// 如果批量生成失败，使用之前生成的单个链接
	txtData := strings.Join(lines, "\n")

	// 写入文件
	return os.WriteFile(*outputPath, []byte(txtData), 0o644)
}
