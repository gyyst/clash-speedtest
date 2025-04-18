package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

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

// ================== Helper Functions ==================
func getStringWithDefault(m map[string]any, keys ...string) string {
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
	if len(keys) > 1 {
		return keys[len(keys)-1]
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
		params.Set("security", "tls")
		if fp := getString(config, "client-fingerprint"); fp != "" {
			params.Set("fp", fp)
		}
		if alpn := getSlice(config, "alpn"); len(alpn) > 0 {
			params.Set("alpn", strings.Join(alpn, ","))
		}

	}
	// Reality协议处理
	if realityOpts, ok := config["reality-opts"].(map[string]any); ok {
		params.Set("security", "reality")
		if pbk := getString(realityOpts, "public-key"); pbk != "" {
			params.Set("pbk", pbk)
			if sid := getString(realityOpts, "short-id"); sid != "" {
				params.Set("sid", sid)
			}
		}
	}
	if sni := getString(config, "servername", "sni"); sni != "" {
		params.Set("sni", sni)
	}
}

func handleWsConfig(config map[string]any, vmess map[string]any) {
	if opts, ok := config["ws-opts"].(map[string]any); ok {
		vmess["path"] = getStringWithDefault(opts, "path", "/")
		if headers, ok := opts["headers"].(map[string]any); ok {
			vmess["host"] = getString(headers, "Host")
		}
	}
}

func handleHttpConfig(config map[string]any, vmess map[string]any) {
	if opts, ok := config["http-opts"].(map[string]any); ok {
		vmess["path"] = getStringWithDefault(opts, "path", "/")
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

func handleTransportParams(config map[string]any, network string, params url.Values) {
	optsKey := network + "-opts"
	if opts, ok := config[optsKey].(map[string]any); ok {
		switch network {
		case "ws":
			if path := getString(opts, "path"); path != "" {
				params.Set("path", path)
			}
			if headers, ok := opts["headers"].(map[string]any); ok {
				if host := getString(headers, "Host"); host != "" {
					params.Set("host", host)
				}
			}
		case "grpc":
			if service := getString(opts, "grpc-service-name"); service != "" {
				params.Set("serviceName", service)
			}
		case "http":
			if path := getString(opts, "path"); path != "" {
				params.Set("path", path)
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
