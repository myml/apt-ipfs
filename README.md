# apt-ipfs
这是一个基于ipfs开发的p2p仓库源加速工具，目前是实验阶段。

ipfs是一个分布式文件系统，底层提供p2p共享能力，类似平常使用的bt下载工具提供p2p，只不过ipfs下载资源不需要提供资源种子，而是需要资源的哈希值。

## 说明
我已经将deepin仓库发布到自己搭建的ipfs节点中进行“做种”，**由于资源刚发布，我自己的ipfs节点资源有限，仓库源访问会比较慢。**

我的ipfs节点ID*已内置配置到工具中*

- 12D3KooWH1d6Zi8WeYbpqaP4MKv23VY6XPXMM4AoSBZq5kv6s4ey
- 12D3KooWDm2o3RZsE7t2oFMqKZxYo4W1c2XwYrKbXm3qXUeVLpnp

deepin仓库CID，由于ipfs是基于资源哈希值寻址，仓库的任何变动都会生成新的CID，建议使用域名作为访问入口

- dnslink: /ipns/mirrors.myml.dev
- 20.2.4版本仓库： /ipfs/QmW2jKhYHRJtcV6Z1c5GKREbn54quwJFZmtHA5jvSRK31G


## 使用

### 安装
以下方式二选一
- Docker安装
```sh
docker run --network host myml/apt-ipfs /apt-ipfs -l 127.0.0.1:8080
# 因为p2p需要节点互连，建议使用主机网络而不是发布端口
```
- 源码编译
```sh
go install github.com/myml/apt-ipfs@latest
~/go/bin/apt-ipfs
```
### 改源
```sh
deb http://127.0.0.1:8080/ipns/mirrors.myml.dev/deepin/ apricot main contrib non-free
deb-src http://127.0.0.1:8080/ipns/mirrors.myml.dev/deepin/ apricot main contrib non-free
```
### 测试
`apt update`
