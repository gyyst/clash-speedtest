package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ================== Shadowsocks ==================
func GenerateShadowsocksLink(proxyName string, config map[string]any) (string, error) {
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

// 将ss格式的节点转换为clash格式
func ParseShadowsocks(data string) (map[string]any, error) {
	if !strings.HasPrefix(data, "ss://") {
		return nil, fmt.Errorf("不是ss格式")
	}
	// 移除 "ss://" 前缀
	data = data[5:]

	// 检查是否包含@分隔符
	if !strings.Contains(data, "@") {
		if strings.Contains(data, "#") {
			temp := strings.SplitN(data, "#", 2)
			data = DecodeBase64(temp[0]) + "#" + temp[1]
		} else {
			data = DecodeBase64(data)
		}
	}
	// 判断是否包含 @ #
	if !strings.Contains(data, "@") && !strings.Contains(data, "#") {
		return nil, fmt.Errorf("格式错误: 缺少@或#分隔符")
	}

	// 分离名称部分
	name := ""
	if idx := strings.LastIndex(data, "#"); idx != -1 {
		name = data[idx+1:]
		name, _ = url.QueryUnescape(name)
		data = data[:idx]
	}

	// 分离用户信息和服务器信息
	parts := strings.SplitN(data, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("格式错误: 缺少@分隔符")
	}

	parts[0] = DecodeBase64(parts[0])

	// 分离加密方式和密码
	methodAndPassword := strings.SplitN(parts[0], ":", 2)
	if len(methodAndPassword) != 2 {
		return nil, fmt.Errorf("格式错误: 加密方式和密码格式不正确")
	}

	method := methodAndPassword[0]

	password := DecodeBase64(methodAndPassword[1])

	// 分离服务器地址和端口
	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) != 2 {
		return nil, fmt.Errorf("格式错误: 服务器地址格式不正确")
	}
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, fmt.Errorf("格式错误: 端口格式不正确")
	}

	// 构建clash格式配置
	proxy := map[string]any{
		"name":     name,
		"type":     "ss",
		"server":   hostPort[0],
		"port":     port,
		"cipher":   method,
		"password": password,
	}

	return proxy, nil
}
