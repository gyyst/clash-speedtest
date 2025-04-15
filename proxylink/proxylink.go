package proxylink

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GenerateProxyLink 主入口函数
func GenerateProxyLink(proxyName string, proxyType string, proxyConfig map[string]any) (string, error) {
	switch strings.ToLower(proxyType) {
	case "vmess":
		return generateVmessLink(proxyName, proxyConfig)
	case "vless":
		return generateVlessLink(proxyName, proxyConfig)
	case "trojan":
		return generateTrojanLink(proxyName, proxyConfig)
	case "shadowsocks", "ss":
		return generateShadowsocksLink(proxyName, proxyConfig)
	case "shadowsocksr", "ssr":
		return generateSSRLink(proxyName, proxyConfig)
	case "hysteria2":
		return generateHysteria2Link(proxyName, proxyConfig)
	default:
		return proxyName, nil
	}
}

// ================== Vmess ==================
func generateVmessLink(proxyName string, config map[string]any) (string, error) {
	base := getBaseParams(config, "uuid")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	vmess := map[string]any{
		"v":    "2",
		"ps":   proxyName,
		"add":  getString(config, "server"),
		"port": getPort(config),
		"id":   getString(config, "uuid"),
		"aid":  getString(config, "alterId", "0"),
		"net":  getString(config, "network", "tcp"),
		"type": "none",
		"tls":  getBool(config, "tls"),
	}
	if getString(config, "cipher") == "auto" {
		vmess["scy"] = "none"
	} else {
		vmess["scy"] = getString(config, "cipher")
	}

	// 处理传输类型
	switch vmess["net"] {
	case "ws":
		handleWsConfig(config, vmess)
	case "http":
		handleHttpConfig(config, vmess)
	case "grpc":
		handleGrpcConfig(config, vmess)
	}

	vmess["sni"] = getString(config, "sni", "servername")
	// TLS处理
	if getBool(config, "tls") {
		vmess["tls"] = "tls"
		vmess["alpn"] = strings.Join(getSlice(config, "alpn"), ",")
		vmess["fp"] = getString(config, "client-fingerprint", "chrome")
	}

	jsonData, _ := json.Marshal(vmess)
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData), nil
}

// ================== Vless ==================
func generateVlessLink(proxyName string, config map[string]any) (string, error) {
	base := getBaseParams(config, "uuid")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	params := url.Values{}
	params.Set("type", getString(config, "network", "tcp"))

	// Flow参数
	if flow := getString(config, "flow"); flow != "" {
		params.Set("flow", flow)
	}

	// TLS处理
	handleTLSConfig(config, params)

	// 传输类型处理
	switch getString(config, "network") {
	case "ws":
		handleWsParams(config, params)
	case "grpc":
		handleGrpcParams(config, params)
	}

	return buildURL("vless", base, proxyName, params), nil
}

// ================== Trojan ==================
func generateTrojanLink(proxyName string, config map[string]any) (string, error) {
	base := getBaseParams(config, "password")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	params := url.Values{}
	if sni := getString(config, "sni", getString(config, "sni")); sni != "" {
		params.Set("sni", sni)
	}
	if alpn := getSlice(config, "alpn"); len(alpn) > 0 {
		params.Set("alpn", strings.Join(alpn, ","))
	}
	// 网络类型处理
	switch network := getString(config, "network"); network {
	case "ws", "grpc", "tcp":
		params.Set("type", network)
		handleTransportParams(config, network, params)
	}

	return buildURL("trojan", base, proxyName, params), nil
}

// ================== Shadowsocks ==================
func generateShadowsocksLink(proxyName string, config map[string]any) (string, error) {
	cipher := getString(config, "cipher")
	password := getString(config, "password")
	server := getString(config, "server")
	port := getPort(config)

	if cipher == "" || password == "" || server == "" || port == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	userInfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cipher, password)))
	params := url.Values{}

	// 插件处理
	if plugin := getString(config, "plugin"); plugin != "" {
		pluginStr := handlePluginOpts(config, plugin)
		params.Set("plugin", pluginStr)
	}

	return buildURL("ss", userInfo+"@"+server+":"+port, proxyName, params), nil
}

// ================== ShadowsocksR ==================
func generateSSRLink(proxyName string, config map[string]any) (string, error) {
	server := getString(config, "server")
	port := getPort(config)
	password := getString(config, "password")
	method := getString(config, "cipher")
	protocol := getString(config, "protocol", "origin")
	obfs := getString(config, "obfs", "plain")

	if server == "" || port == "" || password == "" || method == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	// 构建基础链接部分
	base := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		server,
		port,
		protocol,
		method,
		obfs,
		base64.RawURLEncoding.EncodeToString([]byte(password)))

	// 构建参数部分
	params := make(map[string]string)

	// 处理混淆参数
	if obfsParam := getString(config, "obfs-param"); obfsParam != "" {
		params["obfsparam"] = base64.RawURLEncoding.EncodeToString([]byte(obfsParam))
	}

	// 处理协议参数
	if protocolParam := getString(config, "protocol-param"); protocolParam != "" {
		params["protoparam"] = base64.RawURLEncoding.EncodeToString([]byte(protocolParam))
	}

	// 添加备注（节点名称）
	params["remarks"] = base64.RawURLEncoding.EncodeToString([]byte(proxyName))

	// 处理分组
	if group := getString(config, "group"); group != "" {
		params["group"] = base64.RawURLEncoding.EncodeToString([]byte(group))
	}

	// 构建参数字符串
	var paramStr string
	if len(params) > 0 {
		parts := make([]string, 0, len(params))
		for k, v := range params {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		paramStr = "/" + strings.Join(parts, "&")
	}

	// 组合完整链接并进行Base64编码
	fullLink := base + paramStr
	return "ssr://" + base64.RawURLEncoding.EncodeToString([]byte(fullLink)), nil
}

// ================== Hysteria2 ==================
func generateHysteria2Link(proxyName string, config map[string]any) (string, error) {
	base := getBaseParams(config, "password")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	params := url.Values{}
	if insecure := getBool(config, "skip-cert-verify"); insecure {
		params.Set("insecure", "1")
	}
	if sni := getString(config, "sni"); sni != "" {
		params.Set("sni", sni)
	}

	// 性能参数
	if up := getString(config, "up"); up != "" {
		params.Set("upmbps", up)
	}
	if down := getString(config, "down"); down != "" {
		params.Set("downmbps", down)
	}

	return buildURL("hysteria2", base, proxyName, params), nil
}

// ================== Helper Functions ==================
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

// getBool 从配置映射中获取布尔值
// 参数:
//   - m: 配置映射
//   - keys: 要查找的键名列表
//
// 返回:
//   - bool: 找到的布尔值，如果未找到则返回false
func getBool(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case bool:
				return v
			case string:
				return strings.EqualFold(v, "true")
			case int:
				return v > 0
			}
		}
	}
	return false
}

func getPort(config map[string]any) string {
	if port := getString(config, "port"); port != "" {
		return port
	}
	if port, ok := config["port"].(int); ok {
		return strconv.Itoa(port)
	}
	return ""
}

func getSlice(m map[string]any, key string) []string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case []string:
			return v
		case []any:
			var res []string
			for _, item := range v {
				res = append(res, fmt.Sprintf("%v", item))
			}
			return res
		}
	}
	return nil
}

func getBaseParams(config map[string]any, authKey string) string {
	server := getString(config, "server")
	port := getPort(config)
	auth := getString(config, authKey)
	if server == "" || port == "" || auth == "" {
		return ""
	}
	return fmt.Sprintf("%s@%s:%s", auth, server, port)
}

// ================== Config Handlers ==================
func handleTLSConfig(config map[string]any, params url.Values) {
	if getBool(config, "tls") {
		params.Set("security", "reality")
		if fp := getString(config, "client-fingerprint"); fp != "" {
			params.Set("fp", fp)
		}
		if alpn := getSlice(config, "alpn"); len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		}
		// Reality协议处理
		if realityOpts, ok := config["reality-opts"].(map[string]any); ok {
			if pbk := getString(realityOpts, "public-key"); pbk != "" {
				params.Set("pbk", pbk)
				if sid := getString(realityOpts, "short-id"); sid != "" {
					params.Set("sid", sid)
				}
			}
		}
	}
	if sni := getString(config, "servername", getString(config, "sni")); sni != "" {
		params.Set("sni", sni)
	}
}

func handleWsConfig(config map[string]any, vmess map[string]any) {
	if opts, ok := config["ws-opts"].(map[string]any); ok {
		vmess["path"] = getString(opts, "path", "/")
		if headers, ok := opts["headers"].(map[string]any); ok {
			vmess["host"] = getString(headers, "Host")
		}
	}
}

func handleHttpConfig(config map[string]any, vmess map[string]any) {
	if opts, ok := config["http-opts"].(map[string]any); ok {
		vmess["path"] = getString(opts, "path", "/")
		if headers, ok := opts["headers"].(map[string]any); ok {
			vmess["host"] = getString(headers, "Host")
		}
	}
}

func handleGrpcConfig(config map[string]any, vmess map[string]any) {
	if opts, ok := config["grpc-opts"].(map[string]any); ok {
		vmess["path"] = getString(opts, "grpc-service-name")
	}
}

func handleWsParams(config map[string]any, params url.Values) {
	if opts, ok := config["ws-opts"].(map[string]any); ok {
		if path := getString(opts, "path"); path != "" {
			params.Set("path", path)
		}
		if headers, ok := opts["headers"].(map[string]any); ok {
			if host := getString(headers, "Host"); host != "" {
				params.Set("host", host)
			}
		}
	}
}

func handleGrpcParams(config map[string]any, params url.Values) {
	if opts, ok := config["grpc-opts"].(map[string]any); ok {
		if service := getString(opts, "grpc-service-name"); service != "" {
			params.Set("serviceName", service)
		}
	}
}

func handleTransportParams(config map[string]any, network string, params url.Values) {
	optsKey := network + "-opts"
	if opts, ok := config[optsKey].(map[string]any); ok {
		switch network {
		case "ws":
			if path := getString(opts, "path"); path != "" {
				params.Set("path", path)
			}
		case "grpc":
			if service := getString(opts, "grpc-service-name"); service != "" {
				params.Set("serviceName", service)
			}
		case "tcp":
			// 处理TCP的参数
			if allowInsecure := getBool(config, "skip-cert-verify"); allowInsecure {
				if allowInsecure {
					params.Set("allowInsecure", "1")
				} else {
					params.Set("allowInsecure", "0")
				}
			}
			if peer := getString(config, "sni"); peer != "" {
				params.Set("peer", peer)
			}
		}
	}
}

func handlePluginOpts(config map[string]any, plugin string) string {
	if opts, ok := config["plugin-opts"].(map[string]any); ok {
		var parts []string
		for k, v := range opts {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		return fmt.Sprintf("%s;%s", plugin, strings.Join(parts, ";"))
	}
	return plugin
}

func buildURL(scheme string, auth string, fragment string, params url.Values) string {
	encodedFragment := url.PathEscape(fragment)
	return fmt.Sprintf("%s://%s?%s#%s", scheme, auth, params.Encode(), encodedFragment)
}
