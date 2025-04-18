package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ================== TUIC ==================
func GenerateTuicLink(proxyName string, config map[string]any) (string, error) {
	// 检查必要参数
	base := getBaseParams(config, "password")
	if base == "" {
		return proxyName, fmt.Errorf("missing required parameters")
	}

	// 构建URL参数
	params := url.Values{}

	// 处理UUID
	if uuid := getString(config, "uuid"); uuid != "" {
		params.Set("uuid", uuid)
	}

	// 处理TLS相关参数
	if sni := getString(config, "sni"); sni != "" {
		params.Set("sni", sni)
	}

	if alpn := getSlice(config, "alpn"); len(alpn) > 0 {
		params.Set("alpn", strings.Join(alpn, ","))
	}

	// 处理证书验证
	if allowInsecure := getBool(config, "skip-cert-verify"); allowInsecure {
		params.Set("allowInsecure", "1")
	} else {
		params.Set("allowInsecure", "0")
	}

	// 处理指纹
	if fp := getString(config, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	// 处理拥塞控制算法
	if cc := getString(config, "congestion-controller"); cc != "" {
		params.Set("congestion-controller", cc)
	}

	// 处理UDP中继模式
	if udpRelayMode := getString(config, "udp-relay-mode"); udpRelayMode != "" {
		params.Set("udp-relay-mode", udpRelayMode)
	}

	// 构建URL
	return buildURL("tuic", base, proxyName, params), nil
}

// 将tuic格式的节点转换为clash格式
func ParseTuic(data string) (map[string]any, error) {
	if !strings.HasPrefix(data, "tuic://") {
		return nil, fmt.Errorf("不是tuic格式")
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
		return nil, fmt.Errorf("格式错误: 主机或端口格式不正确")
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
		"name":             name,
		"type":             "tuic",
		"server":           hostPort[0],
		"port":             port,
		"password":         password,
		"uuid":             params.Get("uuid"),
		"skip-cert-verify": params.Get("allowInsecure") == "1",
	}

	// 添加SNI配置
	if sni := params.Get("sni"); sni != "" {
		proxy["sni"] = sni
	}

	// 添加ALPN配置
	if alpn := params.Get("alpn"); alpn != "" {
		proxy["alpn"] = strings.Split(alpn, ",")
	}

	// 添加指纹配置
	if fp := params.Get("fp"); fp != "" {
		proxy["client-fingerprint"] = fp
	}

	// 添加拥塞控制算法
	if cc := params.Get("congestion-controller"); cc != "" {
		proxy["congestion-controller"] = cc
	}

	// 添加UDP中继模式
	if udpRelayMode := params.Get("udp-relay-mode"); udpRelayMode != "" {
		proxy["udp-relay-mode"] = udpRelayMode
	}

	return proxy, nil
}
