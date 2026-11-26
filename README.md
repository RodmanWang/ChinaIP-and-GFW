# ChinaIP-and-GFW

个人自用项目，基于中国 IP/CIDR 列表生成 MaxMind GeoIP2 Country 格式的 `Country.mmdb`。

## 构建

需要 Go 1.23 或更新的 Go 1.x 工具链：

```bash
go mod download
./build.sh
```

构建产物位于 `dist/`：

- `dist/ipip2mmdb`：CIDR 到 MMDB 的生成器
- `dist/verify_ip`：MMDB 查询验证工具

## 生成数据库

```bash
./dist/ipip2mmdb -s ./ChinaIP.txt -d ./Country.mmdb
```

输入文件每行一个 IPv4 或 IPv6 CIDR。空行和以 `#` 开头的整行注释会忽略；非空且格式错误的 CIDR 会使本次生成失败，避免静默产出缺失网段的数据库。生成成功后通过同目录临时文件原子替换目标文件。

## 验证数据库

```bash
./dist/verify_ip ./Country.mmdb
./dist/verify_ip ./Country.mmdb 1.1.1.1 123.126.55.41
```

不传 IP 参数时使用内置示例地址；传入参数时只查询指定地址。

## 版本与兼容性

源码版本：`1.1.0`。数据库类型、国家记录内容、默认输入/输出文件名及原有命令行参数保持兼容。

## 数据来源致谢

- https://github.com/JMVoid/ipip2mmdb
- https://github.com/Hackl0us/GeoIP2-CN
- https://github.com/Masaiki/GeoIP2-CN
- https://github.com/17mon/china_ip_list
- https://github.com/metowolf/iplist
- https://github.com/ChanthMiao/China-IPv6-List
- https://github.com/gaoyifan/china-operator-ip
- https://github.com/felixonmars/dnsmasq-china-list
- https://github.com/Loyalsoldier/v2ray-rules-dat
- https://github.com/misakaio/chnroutes2
