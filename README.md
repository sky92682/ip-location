# 功能特点
-  支持IPv4/IPv6归属地查询；
-  输入域名时自动解析为IP再查询；
-  支持从URL中自动提取IP或域名，直接粘贴完整URL亦可正常查询；
-  无输入时默认显示访问者的IP信息；
-  根据客户端类型，自适应显示效果。  

缺点：使用GeoLite2数据库，精准度稍有不足。

# 效果截图
![image](https://raw.githubusercontent.com/sky92682/ip-location/refs/heads/main/screensnap.png)

# 运行参数
> -host    指定监听的地址（默认为localhost）  
> -port    指定监听的端口（默认为8080）  
> -cityDBPath  指定IP城市数据库的路径（默认为当前目录下的"GeoLite2-City.mmdb"文件）  
> -asnDBPath  指定ASN/运营商数据库的路径（默认为当前目录下的"GeoLite2-ASN.mmdb"文件）  

运行后使用Web浏览器访问即可（使用HTTP协议，若需使用HTTPS，建议通过Caddy等工具反代）。
