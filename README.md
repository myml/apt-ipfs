# apt-ipfs

这是一个基于 ipfs 开发的 p2p 仓库源加速工具，目前是实验阶段。

ipfs 是一个分布式文件系统，底层提供 p2p 共享能力，类似平常使用的 bt 下载工具提供 p2p，只不过 ipfs 下载资源不需要提供资源种子，而是需要资源的哈希值。

## 说明

我已经将 deepin 仓库发布到自己搭建的 ipfs 节点中进行“做种”，**ipfs 节点资源有限，下载会比较慢，使用的人多了就速度就快了。**

我的 ipfs 节点 ID*已内置配置到工具中*

- 12D3KooWH1d6Zi8WeYbpqaP4MKv23VY6XPXMM4AoSBZq5kv6s4ey
- 12D3KooWDm2o3RZsE7t2oFMqKZxYo4W1c2XwYrKbXm3qXUeVLpnp

deepin 仓库 CID，由于 ipfs 是基于资源内容哈希值寻址，仓库的任何变动都会生成新的 CID，建议使用域名（dnslink）作为访问入口

- dnslink: /ipns/mirrors.myml.dev/deepin
- 20.2.4 版本仓库： /ipfs/QmQzNu2pZGyjbmikEbVBJSLSFfFfJ8v8R9z78bHefk2wD2/deepin
- 20.3 版本仓库：/ipfs/QmeZRpcRx4PW6tbRfQuHjjy8U5pgT4ZM67yhGzNXdWwHNy/deepin

## 使用

### 安装

以下方式二选一

- Docker 安装

```sh
# 因为p2p需要节点互连，建议使用主机网络而不是发布端口
docker run -d --name apt-ipfs --network host --restart always -v apt-ipfs-data:/data ghcr.io/myml/apt-ipfs:main /apt-ipfs -l 127.0.0.1:8080
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
