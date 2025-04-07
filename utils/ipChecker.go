package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"

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

	initialResp, err := ic.Client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching Scamalytics risk (initial request): %s\n", err)
		return "", err
	}
	defer initialResp.Body.Close()

	initialBody, err := io.ReadAll(initialResp.Body)

	riskScore := ParseScamalytics(string(initialBody))

	riskScoreInt := 0
	fmt.Sscanf(riskScore, "%d", &riskScoreInt)

	// 判断风险级别
	riskLevel := ""
	if riskScoreInt < 33 {
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
