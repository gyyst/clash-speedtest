package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// 将clash格式的节点转换为trojan格式
func ParseTrojan(proxyName string,data map[string]any) (string, error) {
	// 检查必要参数
	server, ok := data["server"].(string)
	if !ok || server == "" {
		return "", fmt.Errorf("缺少必要参数: server")
	}

	var port int
	switch v := data["port"].(type) {
	case int:
		port = v
	case float64:
		port = int(v)
	case string:
		var err error
		port, err = strconv.Atoi(v)
		if err != nil {
			return "", fmt.Errorf("端口格式不正确")
		}
	default:
		return "", fmt.Errorf("缺少必要参数: port")
	}

	password, ok := data["password"].(string)
	if !ok || password == "" {
		return "", fmt.Errorf("缺少必要参数: password")
	}

	// 构建查询参数
	params := url.Values{}

	// 处理网络类型
	if network, ok := data["network"].(string); ok && network != "" && network != "original" {
		params.Set("type", network)

		// 处理ws配置
		if network == "ws" {
			if wsOpts, ok := data["ws-opts"].(map[string]any); ok {
				if path, ok := wsOpts["path"].(string); ok && path != "" {
					params.Set("path", path)
				}

				if headers, ok := wsOpts["headers"].(map[string]string); ok {
					if host, ok := headers["Host"]; ok && host != "" {
						params.Set("host", host)
					}
				}
			}
		}

		// 处理grpc配置
		if network == "grpc" {
			if grpcOpts, ok := data["grpc-opts"].(map[string]string); ok {
				if serviceName, ok := grpcOpts["serviceName"]; ok && serviceName != "" {
					params.Set("serviceName", serviceName)
				}
			}
		}
	}

	// 处理TLS配置
	if tls, ok := data["tls"].(bool); ok && tls {
		params.Set("security", "tls")

		// 处理SNI
		if sni, ok := data["sni"].(string); ok && sni != "" {
			params.Set("sni", sni)
		} else if sni, ok := data["servername"].(string); ok && sni != "" {
			params.Set("sni", sni)
		}

		// 处理跳过证书验证
		if skipCertVerify, ok := data["skip-cert-verify"].(bool); ok && skipCertVerify {
			params.Set("allowInsecure", "1")
		}

		// 处理ALPN
		if alpnList, ok := data["alpn"].([]string); ok && len(alpnList) > 0 {
			params.Set("alpn", strings.Join(alpnList, ","))
		}
	}

	// 构建完整URL
	resultURL := url.URL{
		Scheme:   "trojan",
		User:     url.User(password),
		Host:     fmt.Sprintf("%s:%d", server, port),
		RawQuery: params.Encode(),
		Fragment: fmt.Sprintf("%v", data["name"]),
	}

	return resultURL.String(), nil
}
