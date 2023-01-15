# nuclei漏洞结果持久化
```
x nuclei -f /Users/leveryd/Downloads/nuclei-result.json -api http://192.168.0.110:30274
```

# 从es http日志中收集子域名
```
./x subdomain -esURL http://192.168.0.110:32116 -domain "apple.com" -of /tmp/result
```

# 从es http日志中识别后台管理系统