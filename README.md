# Clash-SpeedTest

基于 Clash/Mihomo 核心的测速工具，快速测试你的节点速度。

Features:

1. 无需额外的配置，直接将 Clash/Mihomo 配置本地文件路径或者订阅地址作为参数传入即可
2. 支持 Proxies 和 Proxy Provider 中定义的全部类型代理节点，兼容性跟 Mihomo 一致
3. 不依赖额外的 Clash/Mihomo 进程实例，单一工具即可完成测试
4. 代码简单而且开源，不发布构建好的二进制文件，保证你的节点安全
5. 支持测试节点延迟、抖动、丢包率、下载速度和上传速度
6. 支持自动检测节点IP信息，包括国家/地区、城市等
7. 支持多种流媒体解锁测试

<img width="1332" alt="image" src="https://github.com/user-attachments/assets/fdc47ec5-b626-45a3-a38a-6d88c326c588">

## 使用方法

# 支持从源码安装，或从 Release 里下载由 Github Action 自动构建的二进制文件

> go install github.com/faceair/clash-speedtest@latest

# 查看帮助

> clash-speedtest -h
> Usage of clash-speedtest:
> -c string
> configuration file path, also support http(s) url
> -f string
> filter proxies by name, use regexp (default ".*")
> -server-url string
> server url for testing proxies (default "https://speed.cloudflare.com")
> -download-size int
> download size for testing proxies (default 50MB)
> -upload-size int
> upload size for testing proxies (default 20MB)
> -timeout duration
> timeout for testing proxies (default 5s)
> -concurrent int
> download concurrent size (default 4)
> -test-concurrent int
> test proxies concurrent size (default 2)
> -output string
> output config file path (default "result.txt")
> -max-latency duration
> filter latency greater than this value (default 800ms)
> -min-download-speed float
> filter speed less than this value(unit: MB/s) (default 5)
> -min-upload-speed float
> filter upload speed less than this value(unit: MB/s) (default 0)
> -max-packet-loss float
> filter packet loss greater than this value(unit: %) (default 0, max 50)
> -fast
> only test latency, skip download and upload speed test
> -limit int
> limit the number of proxies in output file, 0 means no limit (default 0)
> -unlock string
> test streaming media unlock, support: netflix|chatgpt|disney|youtube|...|all (default:null)
> -sort string
> sort proxies by fields, support: latency|jitter|packet_loss|download|upload, multiple fields separated by | (default "")

# 演示：

# 1. 测试全部节点，使用 HTTP 订阅地址

# 请在订阅地址后面带上 flag=meta 参数，否则无法识别出节点类型

> clash-speedtest -c 'https://domain.com/api/v1/client/subscribe?token=secret&flag=meta'

# 2. 测试香港节点，使用正则表达式过滤，使用本地文件

> clash-speedtest -c ~/.config/clash/config.yaml -f 'HK|港'
> 节点                                        	带宽          	延迟
> Premium|广港|IEPL|01                        	484.80KB/s  	815.00ms
> Premium|广港|IEPL|02                        	N/A         	N/A
> Premium|广港|IEPL|03                        	2.62MB/s    	333.00ms
> Premium|广港|IEPL|04                        	1.46MB/s    	272.00ms
> Premium|广港|IEPL|05                        	3.87MB/s    	249.00ms

# 3. 当然你也可以混合使用

> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta,/home/.config/clash/config.yaml"

# 4. 筛选出延迟低于 800ms 且下载速度大于 5MB/s 的节点，并输出到 filtered.yaml

> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.yaml -max-latency 800ms -min-download-speed 5

# 5. 筛选出延迟低于 800ms、下载速度大于 5MB/s 且上传速度大于 2MB/s 的节点，并输出到 filtered.txt

> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -output filtered.txt -max-latency 800ms -min-download-speed 5 -min-upload-speed 2

filtered.txt示例

```markdown
hysteria2://oejdsadsaMk4AsD4@x.x.x.92:4376?insecure=1&sni=bing.com#🇺🇸美国 SanJose [62% 一般] ...
vless://8adsab-dds9-40cf-802e-70adsa2@14.211.134.145:8080?host=JP.xxxxx.xxxxx.oRg.&path=/?ed=2048&type=ws#🇫🇷 法国 [10%] 纯净 ...
```

# 6. 按照延迟和下载速度排序节点

> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -sort "latency|download"

# 7. 测试节点的流媒体解锁情况

-unlock "all" 则是测试全部

> clash-speedtest -c "https://domain.com/api/v1/client/subscribe?token=secret&flag=meta" -unlock "netflix|disney|youtube"

# 筛选后的配置文件可以直接粘贴到 Clash/Mihomo 中使用，或是贴到 Github\Gist 上通过 Proxy Provider 引用。

## 测速原理

通过 HTTP GET 请求下载指定大小的文件，默认使用 https://speed.cloudflare.com (50MB) 进行测试，计算下载时间得到下载速度。

测试结果：

1. 带宽 是指下载指定大小文件的速度，即一般理解中的下载速度。当这个数值越高时表明节点的出口带宽越大。
2. 延迟 是指 HTTP GET 请求拿到第一个字节的的响应时间，即一般理解中的 TTFB。当这个数值越低时表明你本地到达节点的延迟越低，可能意味着中转节点有 BGP 部署、出海线路是 IEPL、IPLC 等。
3. 抖动 是指多次测试延迟时的波动情况，数值越低表示连接越稳定。
4. 丢包率 是指测试过程中丢失的数据包百分比，数值越低表示连接质量越好。

请注意带宽跟延迟是两个独立的指标，两者并不关联：

1. 可能带宽很高但是延迟也很高，这种情况下你下载速度很快但是打开网页的时候却很慢，可能是是中转节点没有 BGP 加速，但出海线路带宽很充足。
2. 可能带宽很低但是延迟也很低，这种情况下你打开网页的时候很快但是下载速度很慢，可能是中转节点有 BGP 加速，但出海线路的 IEPL、IPLC 带宽很小。

Cloudflare 是全球知名的 CDN 服务商，其提供的测速服务器到海外绝大部分的节点速度都很快，一般情况下都没有必要自建测速服务器。

如果你不想使用 Cloudflare 的测速服务器，可以自己搭建一个测速服务器。

```shell
# 在您需要进行测速的服务器上安装和启动测速服务器
> go install github.com/faceair/clash-speedtest/download-server@latest
> download-server

# 此时在本地使用 http://your-server-ip:8080 作为 server-url 即可
> clash-speedtest --server-url "http://your-server-ip:8080"
```

## IP信息检测

工具会自动检测节点的IP信息，包括：

1. 国家/地区：显示节点所在的国家或地区，并附带国旗emoji
2. 城市：显示节点所在的城市
3. 风险评估：检测IP是否存在风险，如代理/VPN检测、数据中心IP等

这些信息可以帮助你更好地了解节点的地理位置和安全性。

## License

[MIT](LICENSE)
