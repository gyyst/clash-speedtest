package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Ping0Data struct {
	Ping0Risk string
	IPType    string
	NativeIP  string
}

type IPChecker struct {
	Client *http.Client
}

// NewIPChecker 创建一个新的IPChecker实例
func NewIPChecker(client *http.Client) *IPChecker {
	// 如果没有提供客户端，则创建一个默认客户端
	return &IPChecker{Client: client}
}

func (ic *IPChecker) FetchScamalytics(ip string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://scamalytics.com/ip/%s", ip), nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return "", err
	}

	// 添加失败重试功能
	var initialBody []byte
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// 重试前等待随机时间
			time.Sleep(time.Duration(rand.Intn(1000)+100) * time.Millisecond)
			// fmt.Printf("重试获取Scamalytics风险值 (尝试 %d/%d)\n", i+1, maxRetries)
		}

		initialResp, err := ic.Client.Do(req)
		if err != nil {
			lastErr = err
			fmt.Printf("Error fetching Scamalytics risk (attempt %d/%d): %s\n", i+1, maxRetries, err)
			continue
		}

		defer initialResp.Body.Close()
		initialBody, err = io.ReadAll(initialResp.Body)
		if err != nil {
			lastErr = err
			fmt.Printf("Error reading response body (attempt %d/%d): %s\n", i+1, maxRetries, err)
			continue
		}

		// 成功获取响应，跳出重试循环
		lastErr = nil
		break
	}

	// 如果所有重试都失败
	if lastErr != nil {
		return "", fmt.Errorf("达到最大重试次数 (%d): %v", maxRetries, lastErr)
	}

	riskScore := ParseScamalytics(string(initialBody))

	riskScoreInt := 0
	fmt.Sscanf(riskScore, "%d", &riskScoreInt)

	// 判断风险级别
	riskLevel := ""
	if riskScoreInt <= 33 {
		riskLevel = "纯净"
	} else if riskScoreInt <= 66 {
		riskLevel = "一般"
	} else {
		riskLevel = "较差"
	}

	// 格式化为 [分数 风险]
	result := fmt.Sprintf("[%d%% %s]", riskScoreInt, riskLevel)
	return result, nil
}

// ParseScamalytics 解析风险值
func ParseScamalytics(html string) (risk string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}

	// 解析风险值（Fraud Score）
	doc.Find("div.score").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Fraud Score") {
			risk = strings.TrimSpace(strings.Split(s.Text(), ": ")[1])
		}
	})
	return risk
}
