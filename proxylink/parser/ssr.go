package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
)

// 将clash格式的节点转换为ssr格式
func ParseSsr(proxyName string,data map[string]any) (string, error) {
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

	cipher, ok := data["cipher"].(string)
	if !ok || cipher == "" {
		return "", fmt.Errorf("缺少必要参数: cipher")
	}

	protocol, ok := data["protocol"].(string)
	if !ok || protocol == "" {
		return "", fmt.Errorf("缺少必要参数: protocol")
	}

	obfs, ok := data["obfs"].(string)
	if !ok || obfs == "" {
		return "", fmt.Errorf("缺少必要参数: obfs")
	}

	// 构建基本部分
	basePart := fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		server,
		port,
		protocol,
		cipher,
		obfs,
		base64.StdEncoding.EncodeToString([]byte(password)))

	// 构建参数部分
	params := url.Values{}

	// 添加混淆参数
	if obfsParam, ok := data["obfs-param"].(string); ok && obfsParam != "" {
		params.Set("obfsparam", base64.StdEncoding.EncodeToString([]byte(obfsParam)))
	}

	// 添加协议参数
	if protoParam, ok := data["protocol-param"].(string); ok && protoParam != "" {
		params.Set("protoparam", base64.StdEncoding.EncodeToString([]byte(protoParam)))
	}

	// 添加备注
	if name, ok := data["name"].(string); ok && name != "" {
		params.Set("remarks", base64.StdEncoding.EncodeToString([]byte(name)))
	}

	// 构建完整链接
	var ssrURL string
	if len(params) > 0 {
		ssrURL = fmt.Sprintf("%s/?%s", basePart, params.Encode())
	} else {
		ssrURL = basePart
	}

	// 进行Base64编码并添加前缀
	return "ssr://" + base64.StdEncoding.EncodeToString([]byte(ssrURL)), nil
}
