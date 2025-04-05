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

	// 获取必要参数
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")
	uuid := getStringValue(config, "uuid")

	if server == "" || port == "" || uuid == "" {
		return proxyName, fmt.Errorf("无法生成vmess链接：缺少必要参数")
	}

	// 创建vmess链接所需的JSON结构
	vmessConfig := map[string]interface{}{
		"v":    "2",
		"ps":   proxyName,
		"add":  server,
		"port": port,
		"id":   uuid,
		"aid":  getStringValue(config, "alterId", "0"),
		"scy":  getStringValue(config, "cipher", "auto"),
		"net":  "tcp",
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "none",
		"sni":  "",
		"alpn": "",
		"fp":   "chrome",
	}

	// 处理网络类型
	network := getStringValue(config, "network", "ws")
	vmessConfig["net"] = network

	// 处理路径
	if network == "ws" {
		vmessConfig["path"] = getStringValue(config, "ws-path", "/")
		// 处理ws headers
		if headers, ok := config["ws-headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok && host != "" {
				vmessConfig["host"] = host
			}
		} else if host := getStringValue(config, "host"); host != "" {
			vmessConfig["host"] = host
		}
	} else if network == "http" {
		vmessConfig["path"] = getStringValue(config, "http-path", "/")
		vmessConfig["type"] = "http"
		if host := getStringValue(config, "host"); host != "" {
			vmessConfig["host"] = host
		}
	} else if network == "grpc" {
		vmessConfig["path"] = getStringValue(config, "grpc-service-name", "")
		if host := getStringValue(config, "host"); host != "" {
			vmessConfig["host"] = host
		}
	}

	// 处理TLS
	if tls, ok := config["tls"].(bool); ok && tls {
		vmessConfig["tls"] = "tls"
		if sni := getStringValue(config, "servername"); sni != "" {
			vmessConfig["sni"] = sni
		} else if sni := getStringValue(config, "sni"); sni != "" {
			vmessConfig["sni"] = sni
		}

		// 处理ALPN
		if alpn, ok := config["alpn"].([]string); ok && len(alpn) > 0 {
			vmessConfig["alpn"] = strings.Join(alpn, ",")
		} else if alpnStr := getStringValue(config, "alpn"); alpnStr != "" {
			vmessConfig["alpn"] = alpnStr
		}

		// 处理指纹
		if fp := getStringValue(config, "client-fingerprint"); fp != "" {
			vmessConfig["fp"] = fp
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

	// 获取必要参数
	uuid := getStringValue(config, "uuid")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")

	if server == "" || port == "" || uuid == "" {
		return proxyName, fmt.Errorf("无法生成vless链接：缺少必要参数")
	}

	// 设置查询参数
	params := url.Values{}

	// 处理网络类型
	network := getStringValue(config, "network", "tcp")
	params.Set("type", network)

	// 处理安全类型
	if tls, ok := config["tls"].(bool); ok && tls {
		params.Set("security", "tls")

		// 处理SNI
		if sni := getStringValue(config, "servername"); sni != "" {
			params.Set("sni", sni)
		} else if sni := getStringValue(config, "sni"); sni != "" {
			params.Set("sni", sni)
		}

		// 处理指纹
		if fp := getStringValue(config, "client-fingerprint"); fp != "" {
			params.Set("fp", fp)
		}

		// 处理ALPN
		if alpn, ok := config["alpn"].([]string); ok && len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		} else if alpnStr := getStringValue(config, "alpn"); alpnStr != "" {
			params.Set("alpn", alpnStr)
		}

		// 处理Reality
		if pbk := getStringValue(config, "public-key"); pbk != "" {
			params.Set("pbk", pbk)
			params.Set("security", "reality")
		}

		if sid := getStringValue(config, "short-id"); sid != "" {
			params.Set("sid", sid)
			params.Set("security", "reality")
		}
	}

	// 处理路径
	if network == "ws" {
		if path := getStringValue(config, "ws-path"); path != "" {
			params.Set("path", path)
		} else if path := getStringValue(config, "path"); path != "" {
			params.Set("path", path)
		}
	} else if network == "grpc" {
		if svcName := getStringValue(config, "grpc-service-name"); svcName != "" {
			params.Set("serviceName", svcName)
		}
	}

	// 处理Host
	if host := getStringValue(config, "host"); host != "" {
		params.Set("host", host)
	} else if headers, ok := config["ws-headers"].(map[string]interface{}); ok {
		if host, ok := headers["Host"].(string); ok && host != "" {
			params.Set("host", host)
		}
	}

	// 处理额外参数
	if ed := getStringValue(config, "ed"); ed != "" {
		params.Set("ed", ed)
	}

	// 构建URL
	u := url.URL{
		Scheme:   "vless",
		User:     url.User(uuid),
		Host:     server + ":" + port,
		RawQuery: params.Encode(),
		Fragment: proxyName,
	}

	return u.String(), nil
}

// generateTrojanLink 生成trojan链接
func generateTrojanLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成trojan链接：配置为空")
	}

	// 获取必要参数
	password := getStringValue(config, "password")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")

	if server == "" || port == "" || password == "" {
		return proxyName, fmt.Errorf("无法生成trojan链接：缺少必要参数")
	}

	// 设置查询参数
	params := url.Values{}

	// 处理SNI
	if sni := getStringValue(config, "sni"); sni != "" {
		params.Set("sni", sni)
	} else if sni := getStringValue(config, "servername"); sni != "" {
		params.Set("sni", sni)
	}

	// 处理网络类型
	if network := getStringValue(config, "network"); network != "" && network != "tcp" {
		params.Set("type", network)

		// 处理路径
		if network == "ws" {
			if path := getStringValue(config, "ws-path"); path != "" {
				params.Set("path", path)
			} else if path := getStringValue(config, "path"); path != "" {
				params.Set("path", path)
			}
		} else if network == "grpc" {
			if svcName := getStringValue(config, "grpc-service-name"); svcName != "" {
				params.Set("serviceName", svcName)
			}
		}
	}

	// 处理Host
	if host := getStringValue(config, "host"); host != "" {
		params.Set("host", host)
	} else if headers, ok := config["ws-headers"].(map[string]interface{}); ok {
		if host, ok := headers["Host"].(string); ok && host != "" {
			params.Set("host", host)
		}
	}

	// 处理指纹
	if fp := getStringValue(config, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	// 构建URL
	u := url.URL{
		Scheme:   "trojan",
		User:     url.User(password),
		Host:     server + ":" + port,
		RawQuery: params.Encode(),
		Fragment: proxyName,
	}

	return u.String(), nil
}

// generateShadowsocksLink 生成shadowsocks链接
func generateShadowsocksLink(proxyName string, config map[string]any) (string, error) {
	if config == nil {
		return proxyName, fmt.Errorf("无法生成shadowsocks链接：配置为空")
	}

	// 获取必要参数
	cipher := getStringValue(config, "cipher")
	password := getStringValue(config, "password")
	server := getStringValue(config, "server")
	port := getStringValue(config, "port")

	if server == "" || port == "" || password == "" || cipher == "" {
		return proxyName, fmt.Errorf("无法生成shadowsocks链接：缺少必要参数")
	}

	// 构建userinfo部分
	userInfo := fmt.Sprintf("%s:%s", cipher, password)
	userInfoBase64 := base64.StdEncoding.EncodeToString([]byte(userInfo))

	// 设置查询参数
	params := url.Values{}

	// 处理插件
	if plugin := getStringValue(config, "plugin"); plugin != "" {
		pluginOpts := getStringValue(config, "plugin-opts")
		if pluginOpts != "" {
			params.Set("plugin", plugin+";"+pluginOpts)
		} else {
			params.Set("plugin", plugin)
		}
	}

	// 构建URL
	u := url.URL{
		Scheme:   "ss",
		User:     url.User(userInfoBase64),
		Host:     server + ":" + port,
		RawQuery: params.Encode(),
		Fragment: proxyName,
	}

	return u.String(), nil
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

	// 设置查询参数
	params := url.Values{}

	// 添加可选参数
	if sni := getStringValue(config, "sni"); sni != "" {
		params.Set("sni", sni)
	} else if sni := getStringValue(config, "servername"); sni != "" {
		params.Set("sni", sni)
	}

	// 处理跳过证书验证
	if insecure, ok := config["skip-cert-verify"].(bool); ok && insecure {
		params.Set("insecure", "1")
	} else if insecure := getStringValue(config, "insecure"); insecure != "" {
		params.Set("insecure", insecure)
	}

	// 处理混淆
	if obfs := getStringValue(config, "obfs"); obfs != "" {
		params.Set("obfs", obfs)
		if obfsPassword := getStringValue(config, "obfs-password"); obfsPassword != "" {
			params.Set("obfs-password", obfsPassword)
		}
	}

	// 处理指纹
	if fp := getStringValue(config, "fingerprint"); fp != "" {
		params.Set("fp", fp)
	} else if fp := getStringValue(config, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	// 处理ALPN
	if alpn, ok := config["alpn"].([]string); ok && len(alpn) > 0 {
		params.Set("alpn", strings.Join(alpn, ","))
	} else if alpnStr := getStringValue(config, "alpn"); alpnStr != "" {
		params.Set("alpn", alpnStr)
	}

	// 处理多端口
	if mport := getStringValue(config, "ports"); mport != "" {
		params.Set("mport", mport)
	}

	// 处理上下行速率
	if up := getStringValue(config, "up"); up != "" {
		params.Set("upmbps", up)
	}

	if down := getStringValue(config, "down"); down != "" {
		params.Set("downmbps", down)
	}

	// 构建URL
	u := url.URL{
		Scheme:   "hysteria2",
		User:     url.User(password),
		Host:     server + ":" + port,
		RawQuery: params.Encode(),
		Fragment: proxyName,
	}

	return u.String(), nil
}
