package proxylink

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GenerateProxyLink 根据代理类型和配置生成对应的代理链接
func GenerateProxyLink(proxyName string, proxyType string, proxyConfig map[string]any) (string, error) {
	switch proxyType {
	case "Vmess":
		return generateVmessLink(proxyName, proxyConfig)
	case "Vless":
		return generateVlessLink(proxyName, proxyConfig)
	case "Trojan":
		return generateTrojanLink(proxyName, proxyConfig)
	case "Shadowsocks":
		return generateShadowsocksLink(proxyName, proxyConfig)
	case "Hysteria2":
		return generateHysteria2Link(proxyName, proxyConfig)
	default:
		// 对于不支持的类型，返回代理名称
		return proxyName, nil
	}
}

// generateVmessLink 生成vmess链接
func generateVmessLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成vmess链接：配置为空")
	}

	// 创建vmess链接所需的JSON结构
	vmessConfig := map[string]interface{}{
		"v":    "2",
		"ps":   proxyName,
		"add":  getStringValue(config, "server"),
		"port": getStringValue(config, "port"),
		"id":   getStringValue(config, "uuid"),
		"aid":  getStringValue(config, "alterId", "0"),
		"scy":  getStringValue(config, "cipher", "none"),
		"net":  getStringValue(config, "network", "ws"),
		"type": "none",
		"host": "",
		"path": getStringValue(config, "ws-path", "/"),
		"tls":  "",
		"sni":  "",
		"alpn": "",
		"fp":   "",
	}

	// 处理TLS
	if tls, ok := config["tls"].(bool); ok && tls {
		vmessConfig["tls"] = "tls"
		if sni, ok := config["servername"].(string); ok && sni != "" {
			vmessConfig["sni"] = sni
		}
	}

	// 处理ws headers
	if headers, ok := config["ws-headers"].(map[string]interface{}); ok {
		if host, ok := headers["Host"].(string); ok && host != "" {
			vmessConfig["host"] = host
		}
	}

	// 转换为JSON
	jsonData, err := json.Marshal(vmessConfig)
	if err != nil {
		return proxyName, fmt.Errorf("生成vmess链接失败：%v", err)
	}

	// Base64编码
	base64Str := base64.StdEncoding.EncodeToString(jsonData)
	return "vmess://" + base64Str, nil
}

// generateVlessLink 生成vless链接
func generateVlessLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成vless链接：配置为空")
	}

	uuid := getStringValue(config, "uuid")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")
	network := getStringValue(config, "network", "tcp")
	path := getStringValue(config, "ws-path", "/")
	// 构建vless链接
	link := fmt.Sprintf("vless://%s@%s:%s?type=%s", uuid, server, port, network)

	// 添加路径
	if network == "ws" && path != "" {
		link += "&path=" + path
	}

	// 添加TLS
	if tls, ok := config["tls"].(bool); ok && tls {
		link += "&security=tls"
		if sni, ok := config["servername"].(string); ok && sni != "" {
			link += "&sni=" + sni
		}
	} else {
		link += "&security=none"
	}

	// 添加备注
	link += "#" + url.PathEscape(proxyName)

	return link, nil
}

// generateTrojanLink 生成trojan链接
func generateTrojanLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成trojan链接：配置为空")
	}

	password := getStringValue(config, "password")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")

	// 构建trojan链接
	link := fmt.Sprintf("trojan://%s@%s:%s", password, server, port)

	// 添加TLS SNI
	if sni, ok := config["sni"].(string); ok && sni != "" {
		link += "?sni=" + sni
	}

	// 添加备注
	link += "#" + url.PathEscape(proxyName)

	return link, nil
}

// generateShadowsocksLink 生成shadowsocks链接
func generateShadowsocksLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成shadowsocks链接：配置为空")
	}

	cipher := getStringValue(config, "cipher")
	password := getStringValue(config, "password")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")

	// 构建userinfo部分
	userInfo := fmt.Sprintf("%s:%s", cipher, password)
	userInfoBase64 := base64.StdEncoding.EncodeToString([]byte(userInfo))

	// 构建ss链接
	link := fmt.Sprintf("ss://%s@%s:%s", userInfoBase64, server, port)

	// 添加备注
	link += "#" + url.PathEscape(proxyName)

	return link, nil
}

// getStringValue 从配置中获取字符串值，如果不存在则返回默认值
func getStringValue(config map[string]any, key string, defaultValue ...string) string {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case int:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%v", v)
		case bool:
			return fmt.Sprintf("%v", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

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
	link := fmt.Sprintf("hysteria2://%s@%s:%s", password, server, port)

	// 添加查询参数
	params := url.Values{}

	// 添加可选参数
	if sni := getStringValue(config, "sni"); sni != "" {
		params.Add("sni", sni)
	}
	// 添加可选参数
	if insecure := getStringValue(config, "insecure"); insecure != "" {
		params.Add("insecure", insecure)
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
