# NexNS CoreDNS Plugin

 <!-- 假装有项目Logo -->

NexNS CoreDNS Plugin 是一个与 NexNS Controller 协同工作的 CoreDNS 插件，旨在提供开箱即用的 DNS 解析服务。与传统 DNS 解决方案不同，NexNS CoreDNS Plugin 通过与 NexNS Controller 集成，实现了更灵活、可扩展且安全的 DNS 解析。

## 特点

- **集成 NexNS Controller**： 所有名称记录都由 NexNS Controller 管理和同步，解决了传统 DNS Zone Transfer 协议的各种限制。
- **动态 DNS 记录管理**： 实现了实时的 DNS 记录管理，使得修改和更新 DNS 记录变得更加简便。
- **源地址过滤**： 支持根据请求源地址返回不同的 DNS 记录，轻松区分返回局域网和互联网查询结果。
- **开箱即用**： 简单易用的配置和安装步骤，使得 NexNS CoreDNS Plugin 能够快速投入生产环境。

## 使用步骤

1. **下载 CoreDNS**：

    ```bash
    git clone https://github.com/coredns/coredns.git
    ```

2. **编辑 `plugin.cfg` ：**

    请确保nexns位于forward、alternate等插件前，否则轮不到nexns来处理就已经返回查询结果。

    ```txt
    log:log
    nexns:github.com/JourneyBean/nexns-coredns-plugin
    forward:forward
    ```

3. **编译**：

    ```bash
    make gen && make
    ```

4. **编辑 `Corefile` ：**

    ```txt
    . {
        log
        # cache # 注意：如果需要基于源地址返回不同记录，请不要前置缓存！NexNS提前将所有记录保存在内存，因此也无需担忧其查询速度。
        nexns {
            controller http://localhost:8000
        }
        cache
        forward . 192.168.1.1:53
    }
    ```

5. **运行**：

    ```bash
    ./coredns -conf Corefile -p 53
    ```

## 使用示例

```
$ dig A example.com. @127.0.0.1 -p 53 +short
1.2.3.4

$ dig A example.com. @192.168.1.123 -p 53 +short
2.3.4.5
```

## 打包

- **Debian/Ubuntu**
- **OpenWRT**
- **Archlinux**

待补充

## 贡献

待补充

## 许可证

MIT许可，详见LICENSE。

 <!-- Thanks GPT-3.5 for helping generate this document. -->
