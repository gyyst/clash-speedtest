package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type VmessJson struct {
	V    any    `json:"v"`
	Ps   string `json:"ps"`
	Add  string `json:"add"`
	Port any    `json:"port"`
	Id   string `json:"id"`
	Aid  any    `json:"aid"`
	Scy  string `json:"scy"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	Tls  string `json:"tls"`
	Sni  string `json:"sni"`
	Alpn string `json:"alpn"`
	Fp   string `json:"fp"`
}

// 将clash格式的节点转换为vmess格式
func ParseVmess(proxyName string,data map[string]any) (string, error) {
	// 检查必要参数
	uuid, ok := data["uuid"].(string)
	if !ok || uuid == "" {
		return "", fmt.Errorf("缺少必要参数: uuid")
	}

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

	// 构建vmess json对象
	vmessInfo := VmessJson{
		V:    "2",
		Ps:   fmt.Sprintf("%v", data["name"]),
		Add:  server,
		Port: port,
		Id:   uuid,
		Net:  fmt.Sprintf("%v", data["network"]),
		Type: "none",
	}

	// 处理alterId
	if alterId, ok := data["alterId"].(int); ok {
		vmessInfo.Aid = alterId
	} else if alterId, ok := data["alterId"].(string); ok {
		vmessInfo.Aid = alterId
	} else {
		vmessInfo.Aid = 0
	}

	// 处理加密方式
	if cipher, ok := data["cipher"].(string); ok && cipher != "" {
		vmessInfo.Scy = cipher
	} else {
		vmessInfo.Scy = "auto"
	}

	// 处理TLS
	if tls, ok := data["tls"].(bool); ok && tls {
		vmessInfo.Tls = "tls"

		// 处理SNI
		if sni, ok := data["servername"].(string); ok && sni != "" {
			vmessInfo.Sni = sni
		}

		// 处理指纹
		if fp, ok := data["client-fingerprint"].(string); ok && fp != "" {
			vmessInfo.Fp = fp
		}

		// 处理ALPN
		if alpnList, ok := data["alpn"].([]string); ok && len(alpnList) > 0 {
			vmessInfo.Alpn = strings.Join(alpnList, ",")
		}
	} else {
		vmessInfo.Tls = "none"
	}

	// 处理传输方式
	switch vmessInfo.Net {
	case "ws":
		if wsOpts, ok := data["ws-opts"].(map[string]any); ok {
			if path, ok := wsOpts["path"].(string); ok {
				vmessInfo.Path = path
			}

			if headers, ok := wsOpts["headers"].(map[string]string); ok {
				if host, ok := headers["Host"]; ok {
					vmessInfo.Host = host
				}
			}
		}
	case "grpc":
		if grpcOpts, ok := data["grpc-opts"].(map[string]string); ok {
			if serviceName, ok := grpcOpts["grpc-service-name"]; ok {
				vmessInfo.Path = serviceName
			}
		}
	}

	// 转换为JSON
	jsonData, err := json.Marshal(vmessInfo)
	if err != nil {
		return "", fmt.Errorf("生成vmess链接失败: %v", err)
	}

	// 生成vmess链接
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData), nil
}
