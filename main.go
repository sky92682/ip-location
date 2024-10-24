package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// 定义全局变量用于存储数据库引用
var cityDB *geoip2.Reader
var asnDB *geoip2.Reader

func main() {
	// 定义命令行参数
	host := flag.String("host", "localhost", "监听地址")
	port := flag.String("port", "8080", "监听端口")
	cityDBPath := flag.String("citydb", "GeoLite2-City.mmdb", "GeoLite2-City数据库文件路径")
	asnDBPath := flag.String("asndb", "GeoLite2-ASN.mmdb", "GeoLite2-ASN数据库文件路径")

	// 解析命令行参数
	flag.Parse()

	// 打开GeoLite2-City数据库
	var err error
	cityDB, err = geoip2.Open(filepath.Clean(*cityDBPath))
	if err != nil {
		log.Fatalf("无法打开GeoLite2-City数据库文件 %s: %v", *cityDBPath, err)
	}
	defer cityDB.Close()

	// 打开GeoLite2-ASN数据库
	asnDB, err = geoip2.Open(filepath.Clean(*asnDBPath))
	if err != nil {
		log.Fatalf("无法打开GeoLite2-ASN数据库文件 %s: %v", *asnDBPath, err)
	}
	defer asnDB.Close()

	// 设置路由
	http.HandleFunc("/", homeHandler) // 处理首页请求

	// 启动服务器
	address := fmt.Sprintf("%s:%s", *host, *port)
	fmt.Printf("服务器正在运行：http://%s\n", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

// 获取客户端的IP地址
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头中获取
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For 可能是一个用逗号分隔的IP列表，取第一个
		ip := strings.Split(forwarded, ",")[0]
		return strings.TrimSpace(ip)
	}

	// 尝试从 X-Real-IP 头中获取
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 如果以上头部没有获取到，则使用 RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// 提取URL中的主机名
func extractHost(input string) (string, error) {
	// 尝试解析输入是否为URL
	parsedURL, err := url.Parse(input)
	if err != nil || parsedURL.Scheme == "" {
		// 如果输入不是有效的URL，直接返回原始输入
		return input, nil
	}
	// 返回主机名部分（去掉端口号）
	host := parsedURL.Host
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}
	return host, nil
}

// 首页处理函数，显示表单和查询结果
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	query := r.URL.Query().Get("ip")

	// 如果没有提供IP，默认使用客户端IP
	if query == "" {
		query = getClientIP(r)
	}

	// 提取URL中的域名或IP地址
	query, err := extractHost(query)
	if err != nil {
		http.Error(w, "无效的URL或IP地址", http.StatusBadRequest)
		return
	}

	// 初始化HTML结果变量
	var result string

	// 如果有IP地址查询参数，执行查询
	if query != "" {
		// 检查输入是否为域名，如果是则解析为IP地址
		ips, err := net.LookupIP(query)
		if err != nil || len(ips) == 0 {
			result = fmt.Sprintf("<p style='color:red;'>无法解析该IP或域名: %s</p>", query)
		} else {
			// 只使用第一个解析到的IP地址
			ip := ips[0]

			// 查询GeoLite2-City数据库
			record, err := cityDB.City(ip)
			if err != nil {
				result = "<p style='color:red;'>查询GeoLite2-City数据库失败</p>"
			} else {
				// 获取地理位置信息
				country := record.Country.Names["zh-CN"]
				city := record.City.Names["zh-CN"]
				latitude := record.Location.Latitude
				longitude := record.Location.Longitude

				// 获取省份信息（第一个 Subdivision）
				var province string
				if len(record.Subdivisions) > 0 {
					province = record.Subdivisions[0].Names["zh-CN"]
				} else {
					province = "未知"
				}

				// 查询GeoLite2-ASN数据库
				asnRecord, err := asnDB.ASN(ip)
				if err != nil {
					result += "<p style='color:red;'>查询GeoLite2-ASN数据库失败</p>"
				} else {
					asn := asnRecord.AutonomousSystemNumber
					organization := asnRecord.AutonomousSystemOrganization

					// 返回查询结果
					result = fmt.Sprintf(`
						<p>IP地址: %s</p>
						<p>国家/地区: %s</p>
						<p>省份: %s</p>
						<p>城市: %s</p>
						<p>纬度: %f</p>
						<p>经度: %f</p>
						<p>ASN: %d</p>
						<p>运营商: %s</p>
					`, ip.String(), country, province, city, latitude, longitude, asn, organization)
				}
			}
		}
	}

	// 生成HTML页面，显示表单和查询结果
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>IP归属地查询</title>
			<style>
				body, html {
					margin: 0;
					padding: 0;
					height: 100%%;
					display: flex;
					justify-content: center;
					align-items: center;
					background-color: #f0f0f0;
					font-family: Arial, sans-serif;
				}
				.container {
					background-color: #fff;
					padding: 20px;
					border-radius: 8px;
					box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
					max-width: 600px;
					width: 100%%;
					box-sizing: border-box;
				}
				h1 {
					text-align: center;
				}
				form {
					display: flex;
					flex-direction: column;
					gap: 10px;
				}
				label {
					font-size: 1.2em;
				}
				input {
					padding: 10px;
					font-size: 1em;
					border: 1px solid #ccc;
					border-radius: 4px;
				}
				button {
					padding: 10px;
					font-size: 1em;
					background-color: #007bff;
					color: white;
					border: none;
					border-radius: 4px;
					cursor: pointer;
				}
				button:hover {
					background-color: #0056b3;
				}
				.result {
					margin-top: 20px;
					padding: 10px;
					background-color: #e9ecef;
					border-radius: 4px;
				}
				hr {
					margin-top: 20px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>IP归属地查询</h1>
				<form action="/" method="get">
					<label for="ip">请输入IP地址或域名:</label>
					<input type="text" id="ip" name="ip" value="%s" required>
					<button type="submit">查询</button>
				</form>
				<hr>
				<div class="result">
					%s
				</div>
			</div>
		</body>
		</html>
	`, query, result)

	// 将HTML输出到响应中
	fmt.Fprint(w, html)
}
