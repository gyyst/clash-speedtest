package parser

import (
	"fmt"
	"net/url"
	"strconv"
)

func ParseHysteria2(proxyName string,data map[string]any) (string, error) {
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

	// 添加多端口支持
	if ports, ok := data["ports"].(string); ok && ports != "" {
		params.Set("mport", ports)
	}

	// 添加混淆
	if obfs, ok := data["obfs"].(string); ok && obfs != "" {
		params.Set("obfs", obfs)
	}

	// 添加混淆密码
	if obfsPassword, ok := data["obfs-password"].(string); ok && obfsPassword != "" {
		params.Set("obfs-password", obfsPassword)
	}

	// 添加SNI
	if sni, ok := data["sni"].(string); ok && sni != "" {
		params.Set("sni", sni)
	}

	// 添加跳过证书验证
	if skipCertVerify, ok := data["skip-cert-verify"].(bool); ok && skipCertVerify {
		params.Set("insecure", "1")
	}

	// 构建完整URL
	resultURL := url.URL{
		Scheme:   "hysteria2",
		User:     url.User(password),
		Host:     fmt.Sprintf("%s:%d", server, port),
		RawQuery: params.Encode(),
		Fragment: fmt.Sprintf("%v", data["name"]),
	}

	return resultURL.String(), nil
}
