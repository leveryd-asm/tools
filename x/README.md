# nuclei漏洞结果持久化
```
x nuclei -f /Users/leveryd/Downloads/nuclei-result.json -api http://192.168.0.110:30274
```

# 从es http日志中收集子域名
```
./x subdomain -esURL http://192.168.0.110:9200 -domain "apple.com" -of /tmp/result
```

# 从console api中收集子域名
```
./x subdomain -action get -source console -q limit=3000 -domain baidu.com
```

# 从mysql数据库中收集子域名
```
./x subdomain -source mysql -datasource 'root:XXX@tcp(192.168.0.110)/cute' -domain apple.com -sql 'limit 1' -of /tmp/subdomains
```

# 识别后台管理系统
```
./x ims -u https://xxx.com
./x ims -if /tmp/1
```

# url截图
```
./x ss -sssUrl http://192.168.0.110:31824 -u https://www.baidu.com -of ~/Downloads/ -ot dir
```

# 从es http日志中识别后台管理系统

# 从文本中提取域名信息
```
./x txt
```