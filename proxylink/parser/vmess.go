package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// ================== Vmess ==================
func GenerateVmessLink(proxyName string, config map[string]any) (string, error) {
	// 检查必要参数
	base := getBaseParams(config, "uuid")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	// 构建基本vmess配置
	vmess := map[string]any{
		"v":    "2",
		"ps":   proxyName,
		"add":  getString(config, "server"),
		"port": getPort(config),
		"id":   getString(config, "uuid"),
		"aid":  getStringWithDefault(config, "alterId", "0"),
		"net":  getStringWithDefault(config, "network", "tcp"),
		"type": "none",
	}

	// 处理加密方式
	if getString(config, "cipher") != "" {
		vmess["scy"] = getString(config, "cipher")
	} else {
		vmess["scy"] = "auto"
	}

	// 处理传输类型
	switch getString(config, "network", "tcp") {
	case "ws":
		handleWsConfig(config, vmess)
	case "http":
		handleHttpConfig(config, vmess)
	case "grpc":
		handleGrpcConfig(config, vmess)
	}

	// 处理SNI
	vmess["sni"] = getString(config, "sni", "servername")

	// TLS处理
	if getBool(config, "tls") {
		vmess["tls"] = "tls"
		// 处理ALPN
		if alpn := getSlice(config, "alpn"); len(alpn) > 0 {
			vmess["alpn"] = strings.Join(alpn, ",")
		}
		// 处理指纹
		vmess["fp"] = getString(config, "client-fingerprint", "chrome")
	} else {
		vmess["tls"] = "none"
	}

	// 序列化并编码
	jsonData, _ := json.Marshal(vmess)
	return "vmess://" + EncodeBase64(string(jsonData)), nil
}

type VmessJson struct {
	V    any    `json:"v"`    // 类型为any
	Ps   string `json:"ps"`   // 类型为string
	Add  string `json:"add"`  // 类型为string
	Port any    `json:"port"` // 类型为any
	Id   string `json:"id"`   // 类型为string
	Aid  any    `json:"aid"`  // 类型为any
	Scy  string `json:"scy"`  // 类型为string
	Net  string `json:"net"`  // 类型为string
	Type string `json:"type"` // 类型为string
	Host string `json:"host"` // 类型为string
	Path string `json:"path"` // 类型为string
	Tls  string `json:"tls"`  // 类型为string
	Sni  string `json:"sni"`  // 类型为string
	Alpn string `json:"alpn"` // 类型为string
	Fp   string `json:"fp"`   // 类型为string
}

// 将vmess格式的节点转换为clash格式
func ParseVmess(data string) (map[string]any, error) {
	if !strings.HasPrefix(data, "vmess://") {
		slog.Debug(fmt.Sprintf("不是vmess格式: %s", data))
		return nil, fmt.Errorf("不是vmess格式")
	}
	// 移除 "vmess://" 前缀
	data = data[8:]

	// 移除 ` 符号，不知道为什么很多节点结尾有这个
	data = strings.ReplaceAll(data, "`", "")

	// base64解码
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		slog.Debug(fmt.Sprintf("base64解码失败: %s", err))
		return nil, err
	}
	// 解析JSON
	var vmessInfo VmessJson
	if err := json.Unmarshal(decoded, &vmessInfo); err != nil {
		slog.Debug(fmt.Sprintf("json解析失败: %s , 原数据：%s", err, string(decoded)))
		return nil, err
	}

	// 处理 port，支持字符串和数字类型
	var port int
	switch v := vmessInfo.Port.(type) {
	case float64:
		port = int(v)
	case string:
		var err error
		port, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("格式错误: 端口格式不正确")
		}
	default:
		return nil, fmt.Errorf("格式错误: 端口格式不正确")
	}

	var aid int
	switch v := vmessInfo.Aid.(type) {
	case float64:
		aid = int(v)
	case string:
		aid, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("格式错误: alterId格式不正确")
		}
	}

	// 构建clash格式配置
	proxy := map[string]any{
		"name":       vmessInfo.Ps,
		"type":       "vmess",
		"server":     vmessInfo.Add,
		"port":       port,
		"uuid":       vmessInfo.Id,
		"alterId":    aid,
		"cipher":     "auto",
		"network":    vmessInfo.Net,
		"tls":        vmessInfo.Tls == "tls",
		"servername": vmessInfo.Sni,
		// 添加原格式
		"raw": vmessInfo,
	}

	// 根据不同传输方式添加特定配置
	switch vmessInfo.Net {
	case "ws":
		wsOpts := map[string]any{
			"path": vmessInfo.Path,
		}
		if vmessInfo.Host != "" {
			wsOpts["headers"] = map[string]string{
				"Host": vmessInfo.Host,
			}
		}
		proxy["ws-opts"] = wsOpts
	case "grpc":
		grpcOpts := map[string]string{
			"serviceName": vmessInfo.Path,
		}
		proxy["grpc-opts"] = grpcOpts
	}

	// 添加 ALPN 配置
	if vmessInfo.Alpn != "" {
		proxy["alpn"] = strings.Split(vmessInfo.Alpn, ",")
	}

	return proxy, nil
}
