package parser

import (
	"fmt"
	"net/url"
	"strconv"
)

// 将clash格式的节点转换为vless格式的节点
func ParseVless(proxyName string,data map[string]any) (string, error) {
	// 检查必要参数
	uuid, ok := data["uuid"].(string)
	if !ok || uuid == "" {
		return "", fmt.Errorf("缺少必要参数: uuid")
	}

	server, ok := data["server"].(string)
	if !ok || server == "" {
		return "", fmt.Errorf("缺少必要参数: server")
	}

	port, ok := data["port"].(int)
	if !ok {
		// 尝试将其他类型转换为int
		portValue, ok := data["port"]
		if !ok {
			return "", fmt.Errorf("缺少必要参数: port")
		}

		switch v := portValue.(type) {
		case float64:
			port = int(v)
		case string:
			var err error
			port, err = strconv.Atoi(v)
			if err != nil {
				return "", fmt.Errorf("端口格式不正确")
			}
		default:
			return "", fmt.Errorf("端口格式不正确")
		}
	}

	// 构建查询参数
	params := url.Values{}

	// 设置网络类型
	network, _ := data["network"].(string)
	if network == "" {
		network = "tcp"
	}
	params.Set("type", network)

	// 设置TLS
	tls, ok := data["tls"].(bool)
	if ok && tls {
		params.Set("security", "tls")

		// 设置SNI
		if sni, ok := data["servername"].(string); ok && sni != "" {
			params.Set("sni", sni)
		}

		// 设置指纹
		if fp, ok := data["client-fingerprint"].(string); ok && fp != "" {
			params.Set("fp", fp)
		}
	} else {
		params.Set("security", "none")
	}

	// 设置流控
	if flow, ok := data["flow"].(string); ok && flow != "" {
		params.Set("flow", flow)
	}

	// 设置UDP
	if udp, ok := data["udp"].(bool); ok && udp {
		params.Set("udp", "true")
	}

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

	// 处理reality配置
	if realityOpts, ok := data["reality-opts"].(map[string]string); ok {
		if pbk, ok := realityOpts["public-key"]; ok && pbk != "" {
			params.Set("pbk", pbk)
		}

		if sid, ok := realityOpts["short-id"]; ok && sid != "" {
			params.Set("sid", sid)
		}
	}

	// 处理grpc配置
	if network == "grpc" {
		if grpcOpts, ok := data["grpc-opts"].(map[string]string); ok {
			if serviceName, ok := grpcOpts["grpc-service-name"]; ok && serviceName != "" {
				params.Set("serviceName", serviceName)
			}
		}
	}

	// 构建完整URL
	resultURL := url.URL{
		Scheme:   "vless",
		User:     url.User(uuid),
		Host:     fmt.Sprintf("%s:%d", server, port),
		RawQuery: params.Encode(),
		Fragment: fmt.Sprintf("%v", data["name"]),
	}

	return resultURL.String(), nil
}
