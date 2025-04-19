package proxylink

import (
	"bytes"
	"strings"

	"github.com/faceair/clash-speedtest/proxylink/parser"
)

// GenerateProxyLink 主入口函数
func GenerateProxyLink(proxyName string, proxyType string, proxyConfig map[string]any) (string, error) {
	switch strings.ToLower(proxyType) {
	case "vmess":
		return parser.GenerateVmessLink(proxyName, proxyConfig)
	case "vless":
		return parser.GenerateVlessLink(proxyName, proxyConfig)
	case "trojan":
		return parser.GenerateTrojanLink(proxyName, proxyConfig)
	case "shadowsocks", "ss":
		return parser.GenerateShadowsocksLink(proxyName, proxyConfig)
	case "shadowsocksr", "ssr":
		return parser.GenerateSSRLink(proxyName, proxyConfig)
	case "hysteria2", "hy2":
		return parser.GenerateHysteria2Link(proxyName, proxyConfig)
	case "tuic", "tuic5":
		return parser.GenerateTuicLink(proxyName, proxyConfig)
	default:
		return proxyName, nil
	}
}

// ParseProxiesJSON 解析JSON格式的代理配置并返回URL格式
func ParseProxiesJSON(data []byte) (*bytes.Buffer, error) {
	return parser.GenUrls(data)
}
