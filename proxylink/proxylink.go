package proxylink

import (
	"strings"

	"github.com/faceair/clash-speedtest/proxylink/parser"
)

// GenerateProxyLink 主入口函数
func GenerateProxyLink(proxyName string, proxyType string, proxyConfig map[string]any) (string, error) {
	switch strings.ToLower(proxyType) {
	case "vmess":
		return parser.ParseVmess(proxyName, proxyConfig)
	case "vless":
		return parser.ParseVless(proxyName, proxyConfig)
	case "trojan":
		return parser.ParseTrojan(proxyName, proxyConfig)
	case "shadowsocks", "ss":
		return parser.ParseShadowsocks(proxyName, proxyConfig)
	case "shadowsocksr", "ssr":
		return parser.ParseSsr(proxyName, proxyConfig)
	case "hysteria2":
		return parser.ParseHysteria2(proxyName, proxyConfig)
	default:
		return proxyName, nil
	}
}
