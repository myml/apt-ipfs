# apt-ipfs

这是一个基于 ipfs 开发的 p2p 仓库源加速工具，目前是实验阶段。

ipfs 是一个分布式文件系统，底层提供 p2p 共享能力，类似平常使用的 bt 下载工具，只不过 ipfs 下载资源不需要提供资源种子，而是需要资源的哈希值。

## 说明

我已经将 deepin 仓库发布到自己搭建的 ipfs 节点中进行“做种”，**节点资源有限，下载会比较慢，使用的人多了就速度就快了。**

我的 ipfs 节点 ID*已内置配置到工具中*

- 12D3KooWQYZMiH1vGpNKXh6jp8XnZ5mKEmFa3G4H5y7JN7KPV7ZF

deepin 仓库 CID，由于 ipfs 是基于资源内容哈希值寻址，仓库的任何变动都会生成新的 CID，建议使用域名（dnslink）作为访问入口

- dnslink: /ipns/mirrors.getdeepin.org/deepin
- 2023-2-11 版本仓库：/ipfs/QmUE3METyy3k6oYofFtReaRcXp4hLfPty2AdrMWkmqgoiF/deepin

## 安装

### 使用 DEB 包安装

到 [Release](https://github.com/myml/apt-ipfs/releases) 页面下载 deb 包安装使用

### 使用 Docker 安装

因为p2p需要节点互连，建议使用主机网络而不是发布端口

```sh
docker run -d --name apt-ipfs --network host --restart always -v apt-ipfs-data:/data ghcr.io/myml/apt-ipfs:main /apt-ipfs -l 127.0.0.1:12380
```

### 从源码安装

```sh
go install github.com/myml/apt-ipfs@latest
```

## 使用

### 改源

```sh
deb http://127.0.0.1:12380/ipns/mirrors.getdeepin.org/deepin/ apricot main contrib non-free
deb-src http://127.0.0.1:12380/ipns/mirrors.getdeepin.org/deepin/ apricot main contrib non-free
```

### 测试

`sudo apt update && apt download wget`
