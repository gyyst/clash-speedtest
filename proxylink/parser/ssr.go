package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ================== ShadowsocksR ==================
func GenerateSSRLink(proxyName string, config map[string]any) (string, error) {
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
func ParseSsr(data string) (map[string]any, error) {
	if !strings.HasPrefix(data, "ssr://") {
		return nil, fmt.Errorf("不是ssr格式")
	}
	// todo: 这些参数解析应该也有问题
	data = strings.TrimPrefix(data, "ssr://")
	data = DecodeBase64(data)
	serverInfoAndParams := strings.SplitN(data, "/?", 2)
	parts := strings.Split(serverInfoAndParams[0], ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("ssr 参数错误")
	}
	server := parts[0]
	protocol := parts[2]
	method := parts[3]
	obfs := parts[4]
	password := DecodeBase64(parts[5])
	portStr := parts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("ssr 端口错误")
	}
	var obfsParam string
	var protoParam string
	var remarks string
	if len(serverInfoAndParams) == 2 {
		params, err := url.ParseQuery(serverInfoAndParams[1])
		if err != nil {
			return nil, fmt.Errorf("ssr 参数错误")
		}
		if params.Get("obfsparam") != "" {
			obfsParam = DecodeBase64(params.Get("obfsparam"))
		}
		if params.Get("protoparam") != "" {
			protoParam = DecodeBase64(params.Get("protoparam"))
		}
		if params.Get("remarks") != "" {
			remarks = DecodeBase64(params.Get("remarks"))
		} else {
			remarks = server + ":" + strconv.Itoa(port)
		}

	}
	return map[string]any{
		"name":           remarks,
		"server":         server,
		"port":           port,
		"password":       password,
		"cipher":         method,
		"obfs":           obfs,
		"obfs-param":     obfsParam,
		"protocol":       protocol,
		"protocol-param": protoParam,
	}, nil
}
