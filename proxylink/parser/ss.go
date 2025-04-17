package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
)

// 将clash格式的节点转换为ss格式
func ParseShadowsocks(proxyName string,data map[string]any) (string, error) {
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

	cipher, ok := data["cipher"].(string)
	if !ok || cipher == "" {
		return "", fmt.Errorf("缺少必要参数: cipher")
	}

	password, ok := data["password"].(string)
	if !ok || password == "" {
		return "", fmt.Errorf("缺少必要参数: password")
	}

	// 构建用户信息部分
	userInfo := fmt.Sprintf("%s:%s", cipher, password)
	encodedUserInfo := base64.StdEncoding.EncodeToString([]byte(userInfo))

	// 构建服务器地址部分
	serverAddress := fmt.Sprintf("%s:%d", server, port)

	// 构建完整的URL
	ssURL := fmt.Sprintf("ss://%s@%s", encodedUserInfo, serverAddress)

	// 添加节点名称
	if name, ok := data["name"].(string); ok && name != "" {
		ssURL = fmt.Sprintf("%s#%s", ssURL, url.QueryEscape(name))
	}

	return ssURL, nil
}
