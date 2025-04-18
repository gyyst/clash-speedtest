package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ================== Trojan ==================
func GenerateTrojanLink(proxyName string, config map[string]any) (string, error) {
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
	// 处理TCP的参数
	if allowInsecure := getBool(config, "skip-cert-verify"); allowInsecure {
		if allowInsecure {
			params.Set("allowInsecure", "1")
		} else {
			params.Set("allowInsecure", "0")
		}
	}
	// 网络类型处理
	switch network := getString(config, "network"); network {
	case "ws", "grpc", "tcp":
		params.Set("type", network)
		handleTransportParams(config, network, params)
	}

	return buildURL("trojan", base, proxyName, params), nil
}

// 将trojan格式的节点转换为clash格式
func ParseTrojan(data string) (map[string]any, error) {
	if !strings.HasPrefix(data, "trojan://") {
		return nil, fmt.Errorf("不是trojan格式")
	}

	// 解析URL
	u, err := url.Parse(data)
	if err != nil {
		return nil, err
	}

	// 提取密码
	password := u.User.String()

	// 分离主机和端口
	hostPort := strings.Split(u.Host, ":")
	if len(hostPort) != 2 {
		return nil, nil
	}

	// 提取节点名称
	name := ""
	if fragment := u.Fragment; fragment != "" {
		name = fragment
	}

	// 解析查询参数
	params := u.Query()
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, fmt.Errorf("格式错误: 端口格式不正确")
	}

	// 构建clash格式配置
	proxy := map[string]any{
		"name":     name,
		"type":     "trojan",
		"server":   hostPort[0],
		"port":     port,
		"password": password,
		"network": func() string {
			if t := params.Get("type"); t != "" {
				return t
			} else {
				return "original"
			}
		}(),
		"skip-cert-verify": params.Get("allowInsecure") == "1",
		// 添加原格式
		"allowInsecure": params.Get("allowInsecure"),
	}

	// 添加TLS配置
	if params.Get("security") == "tls" {
		proxy["tls"] = true
		if sni := params.Get("sni"); sni != "" {
			proxy["sni"] = sni
		}
	}

	// 根据不同传输方式添加特定配置
	switch params.Get("type") {
	case "ws":
		wsOpts := map[string]any{
			"path": params.Get("path"),
		}
		if host := params.Get("host"); host != "" {
			wsOpts["headers"] = map[string]string{
				"Host": host,
			}
		}
		proxy["ws-opts"] = wsOpts
	case "grpc":
		if serviceName := params.Get("serviceName"); serviceName != "" {
			proxy["grpc-opts"] = map[string]string{
				"serviceName": serviceName,
			}
		}
	}

	return proxy, nil
}
