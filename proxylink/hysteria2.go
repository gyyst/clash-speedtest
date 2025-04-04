package proxylink

import (
	"fmt"
	"net/url"
	"strings"
)

// generateHysteria2Link 生成hysteria2链接
func generateHysteria2Link(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成hysteria2链接：配置为空")
	}

	// 获取必要参数
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")
	password := getStringValue(config, "password")

	if server == "" || port == "" || password == "" {
		return proxyName, fmt.Errorf("无法生成hysteria2链接：缺少必要参数")
	}

	// 构建基本URL
	link := fmt.Sprintf("hy2://%s@%s:%s", password, server, port)

	// 添加查询参数
	params := url.Values{}

	// 添加可选参数
	if sni := getStringValue(config, "sni"); sni != "" {
		params.Add("sni", sni)
	}

	if obfs := getStringValue(config, "obfs"); obfs != "" {
		params.Add("obfs", obfs)
		if obfsPassword := getStringValue(config, "obfs-password"); obfsPassword != "" {
			params.Add("obfs-password", obfsPassword)
		}
	}

	if fingerprint := getStringValue(config, "fingerprint"); fingerprint != "" {
		params.Add("fp", fingerprint)
	}

	if alpn, ok := config["alpn"].([]string); ok && len(alpn) > 0 {
		params.Add("alpn", strings.Join(alpn, ","))
	} else if alpnStr := getStringValue(config, "alpn"); alpnStr != "" {
		params.Add("alpn", alpnStr)
	}

	// 添加查询参数到链接
	if len(params) > 0 {
		link += "?" + params.Encode()
	}

	// 添加备注
	link += "#" + url.PathEscape(proxyName)

	return link, nil
}
