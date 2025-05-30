package parser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
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
		if serviceName := getString(opts, "grpc-service-name"); serviceName != "" {
			vmess["path"] = serviceName
		} else if serviceName := getString(opts, "serviceName"); serviceName != "" {
			vmess["path"] = serviceName
		}
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
	if params.Encode() == "" {
		return fmt.Sprintf("%s://%s#%s", scheme, auth, encodedFragment)
	}
	return fmt.Sprintf("%s://%s?%s#%s", scheme, auth, params.Encode(), encodedFragment)
}

// 生成类似urls
// hysteria2://b82f14be-9225-48cb-963e-0350c86c31d3@us2.interld123456789.com:32000/?insecure=1&sni=234224.1234567890spcloud.com&mport=32000-33000#美国hy2-2-联通电信
// hysteria2://b82f14be-9225-48cb-963e-0350c86c31d3@sg1.interld123456789.com:32000/?insecure=1&sni=234224.1234567890spcloud.com&mport=32000-33000#新加坡hy2-1-移动优化
// 被我拉成屎山了，因为从yaml解析成URI很累很累，这里很多不规范
func GenUrls(data []byte) (*bytes.Buffer, error) {
	urls := bytes.NewBuffer(make([]byte, 0, len(data)*11/10))

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		name, err := jsonparser.GetString(value, "name")
		if err != nil {
			slog.Debug(fmt.Sprintf("获取name字段失败: %s", err))
			return
		}

		// 获取必需字段
		t, err := jsonparser.GetString(value, "type")
		if err != nil {
			slog.Debug(fmt.Sprintf("获取type字段失败: %s", err))
			return
		}

		password, err := jsonparser.GetString(value, "password")
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				password, _ = jsonparser.GetString(value, "uuid")
			} else {
				slog.Debug(fmt.Sprintf("获取password/uuid字段失败: %s", err))
				return
			}
		}
		// 如果是ss，则将cipher和password拼接
		if t == "ss" {
			cipher, err := jsonparser.GetString(value, "cipher")
			if err != nil {
				slog.Debug(fmt.Sprintf("获取cipher字段失败: %s", err))
				return
			}
			password = base64.StdEncoding.EncodeToString([]byte(cipher + ":" + password))
		}
		server, err := jsonparser.GetString(value, "server")
		if err != nil {
			slog.Debug(fmt.Sprintf("获取server字段失败: %s", err))
			return
		}
		port, err := jsonparser.GetInt(value, "port")
		if err != nil {
			slog.Debug(fmt.Sprintf("获取port字段失败: %s", err))
			return
		}

		if t == "ssr" {
			// ssr://host:port:protocol:method:obfs:urlsafebase64pass/?obfsparam=urlsafebase64&protoparam=&remarks=urlsafebase64&group=urlsafebase64&udpport=0&uot=1
			protocol, _ := jsonparser.GetString(value, "protocol")
			cipher, _ := jsonparser.GetString(value, "cipher")
			obfs, _ := jsonparser.GetString(value, "obfs")
			password = base64.URLEncoding.EncodeToString([]byte(password))
			name = base64.URLEncoding.EncodeToString([]byte(name))
			obfsParam, _ := jsonparser.GetString(value, "obfs-param")
			protoParam, _ := jsonparser.GetString(value, "protocol-param")

			url := server + ":" + strconv.Itoa(int(port)) + ":" + protocol + ":" + cipher + ":" + obfs + ":" + password + "/?obfsparam=" + base64.URLEncoding.EncodeToString([]byte(obfsParam)) + "&protoparam=" + base64.URLEncoding.EncodeToString([]byte(protoParam)) + "&remarks=" + name

			urls.WriteString("ssr://" + base64.StdEncoding.EncodeToString([]byte(url)))
			urls.WriteByte('\n')
			return
		}
		// 如果是vmess，则将raw字段base64编码，直接返回
		if t == "vmess" {
			raw, _, _, err := jsonparser.Get(value, "raw")
			if err != nil {
				if err != jsonparser.KeyPathNotFoundError {
					slog.Debug(fmt.Sprintf("获取raw字段失败: %s", err))
					return
				}

				aid, _ := jsonparser.GetInt(value, "aid")
				network, _ := jsonparser.GetString(value, "network")
				tls, _ := jsonparser.GetBoolean(value, "tls")
				servername, _ := jsonparser.GetString(value, "servername")
				alpn, _, _, _ := jsonparser.Get(value, "alpn")
				host, _ := jsonparser.GetString(value, "ws-opts", "headers", "Host")
				path, _ := jsonparser.GetString(value, "ws-opts", "path")
				vmess := VmessJson{
					V:    "2",
					Ps:   name,
					Add:  server,
					Port: port,
					Id:   password,
					Aid:  aid,
					Scy:  "auto",
					Net:  network,
					Type: func() string {
						if network == "http" {
							return "http"
						} else {
							return ""
						}
					}(),
					Host: host,
					Path: path,
					Tls: func() string {
						if tls {
							return "tls"
						} else {
							return "none"
						}
					}(),
					Sni:  servername,
					Alpn: string(alpn),
					Fp:   "chrome",
				}
				d, _ := json.Marshal(vmess)
				urls.WriteString("vmess://")
				urls.WriteString(base64.StdEncoding.EncodeToString(d))
				urls.WriteByte('\n')
				return
			}
			// 因为vmess是json格式，前边的重命名对这里边不起作用，这里单独处理
			raw, err = jsonparser.Set(raw, []byte(fmt.Sprintf(`"%s"`, name)), "ps")
			if err != nil {
				slog.Debug(fmt.Sprintf("修改vmess ps字段失败: %s", err))
				return
			}
			urls.WriteString("vmess://")
			urls.WriteString(base64.StdEncoding.EncodeToString(raw))
			urls.WriteByte('\n')
			return
		}

		// 设置查询参数
		q := url.Values{}

		// 检测vless 如果开了tls，则设置security为tls,后边如果发现有sid字段，则设置security为reality
		tls, _ := jsonparser.GetBoolean(value, "tls")
		if tls {
			q.Set("security", "tls")
		}
		err = jsonparser.ObjectEach(value, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
			keyStr := string(key)
			// 跳过已处理的基本字段
			switch keyStr {
			case "type", "password", "server", "port", "name", "uuid":
				return nil

			// 单独处理vless，因为vless的clash的network字段是url的type字段
			// 我也不知道有没有更好的正确的处理方法或者库
			case "network":
				if t == "vless" {
					q.Set("type", string(val))
				}
				return nil
			}

			// 将clash的参数转换为url的参数
			conversion := func(k, v string) {
				if v == "" {
					return
				}
				switch k {
				case "servername":
					if t == "hysteria" {
						q.Set("peer", v)
					} else {
						q.Set("sni", v)
					}
				case "client-fingerprint":
					q.Set("fp", v)
				case "public-key":
					q.Set("pbk", v)
				case "short-id":
					q.Set("sid", v)
					q.Set("security", "reality")
				case "ports":
					q.Set("mport", v)
				case "skip-cert-verify":
					if v == "true" {
						q.Set("insecure", "1")
						q.Set("allowInsecure", "1")
					} else {
						q.Set("insecure", "0")
						q.Set("allowInsecure", "0")
					}
				case "Host":
					q.Set("host", v)
				case "grpc-service-name":
					q.Set("serviceName", v)
				// hysteria 用的
				case "down":
					q.Set("downmbps", v)
				case "up":
					q.Set("upmbps", v)
				case "auth_str":
					q.Set("auth", v)

				default:
					q.Set(k, v)
				}
			}

			// 如果val是对象，则递归解析
			if dataType == jsonparser.Object {
				return jsonparser.ObjectEach(val, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
					// vless的特殊情况 headers {"host":"vn.oldcloud.online"}
					// 前边处理过vless了，暂时保留，万一后边其他协议还需要
					if dataType == jsonparser.Object {
						return jsonparser.ObjectEach(val, func(key []byte, val []byte, dataType jsonparser.ValueType, offset int) error {
							conversion(string(key), string(val))
							return nil
						})
					}
					conversion(string(key), string(val))
					return nil
				})
			} else {
				conversion(keyStr, string(val))
			}

			return nil
		})
		if err != nil {
			slog.Debug(fmt.Sprintf("获取其他字段失败: %s", err))
			return
		}

		u := url.URL{
			Scheme:   t,
			User:     url.User(password),
			Host:     server + ":" + strconv.Itoa(int(port)),
			RawQuery: q.Encode(),
			Fragment: name,
		}
		if t == "hysteria" {
			u = url.URL{
				Scheme:   t,
				Host:     server + ":" + strconv.Itoa(int(port)),
				RawQuery: q.Encode(),
				Fragment: name,
			}
		}
		urls.WriteString(u.String())
		urls.WriteByte('\n')
	})

	if err != nil {
		return nil, fmt.Errorf("解析代理配置转成urls时失败: %w", err)
	}

	return urls, nil
}
