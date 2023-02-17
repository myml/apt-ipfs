package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/corehttp"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ipfs/kubo/plugin/loader"

	"github.com/coreos/go-systemd/v22/activation"
)

var (
	Listen   = ":8080"
	RepoPath = "./data"

	Mirrors = "/ipns/mirrors.getdeepin.org"
	HotData = "/ipns/mirrors.getdeepin.org/.hotdata"

	Peers     = "_mirrors.getdeepin.org"
	SwarmKey  = "/key/swarm/psk/1.0.0/\n/base16/\n2508242b6ac9e665ea98eb134dd7e05497530f36876ae3ebc865f45fb104291b"
	Bootstrap = []string{
		"/dns4/bootstrap.getdeepin.org/tcp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
		"/dns4/bootstrap.getdeepin.org/udp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
		"/dns6/bootstrap.getdeepin.org/tcp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
		"/dns6/bootstrap.getdeepin.org/udp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
	}
)

func main() {
	flag.StringVar(&RepoPath, "p", RepoPath, "ipfs repo path")
	flag.StringVar(&Listen, "l", Listen, "listen address")
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	node, err := initNode(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("inited")
	go pinHotData(ctx, node)
	var l net.Listener
	list, err := activation.Listeners()
	if err != nil {
		log.Fatal(err)
	}
	if len(list) > 0 {
		l = list[0]
	} else {
		l, err = net.Listen("tcp", Listen)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("listen: ", Listen)
	}
	opts := []corehttp.ServeOption{
		corehttp.GatewayOption(false, "/ipfs", "/ipns"),
	}
	err = corehttp.Serve(node, l, opts...)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exiting")
	time.Sleep(time.Second)
}

// 定时缓存热点数据
func pinHotData(ctx context.Context, node *core.IpfsNode) {
	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		log.Fatal(err)
	}
	isFirstTime := true
	for {
		if isFirstTime {
			isFirstTime = false
		} else {
			time.Sleep(time.Minute)
		}
		peers, err := api.Swarm().Peers(ctx)
		if err != nil {
			log.Println("get current peers:", err)
			continue
		}
		stats := node.Reporter.GetBandwidthTotals()
		log.Printf("Peers count: %v\tTotal Up: %v\tTotal Down: %v\n", len(peers), humanize.Bytes(uint64(stats.TotalOut)), humanize.Bytes(uint64(stats.TotalIn)))

		hotCid, err := resolveIPNS(ctx, node, HotData)
		if err != nil {
			log.Println("resolve hotdata:", err)
			continue
		}
		path := path.New(HotData)
		exists := false
		ch, err := api.Pin().Ls(ctx, options.Pin.Ls.Recursive())
		for info := range ch {
			if !info.Path().Cid().Equals(*hotCid) {
				err = api.Pin().Rm(ctx, info.Path())
				log.Println("resolve hotdata:", err)
				if err != nil {
					log.Println("rm pin:", err)
					continue
				}
			} else {
				exists = true
			}
		}
		if !exists {
			log.Println("pin hot data", hotCid, api.Pin().Add(ctx, path, options.Pin.Recursive(true)))
		}
	}
}

// 根据ipns获取cid
func resolveIPNS(ctx context.Context, node *core.IpfsNode, ipns string) (*cid.Cid, error) {
	core, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, fmt.Errorf("core api:%w", err)
	}
	p, err := core.Name().Resolve(ctx, ipns)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	info, err := core.Object().Stat(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("parse cid: %w", err)
	}
	return &info.Cid, nil
}

// 初始化节点
func initNode(ctx context.Context) (*core.IpfsNode, error) {
	plugins, err := loader.NewPluginLoader(filepath.Join(RepoPath, "plugins"))
	if err != nil {
		return nil, fmt.Errorf("loading plugins: %w", err)
	}
	if err := plugins.Initialize(); err != nil {
		return nil, fmt.Errorf("initializing plugins: %w", err)
	}
	if err := plugins.Inject(); err != nil {
		return nil, fmt.Errorf("initializing plugins: %w", err)
	}
	if _, err = os.Stat(filepath.Join(RepoPath, "config")); errors.Is(err, os.ErrNotExist) {
		_, err = initConfig()
		if err != nil {
			return nil, err
		}
	}
	repo, err := fsrepo.Open(RepoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	nodeOptions := &core.BuildCfg{Online: true, Repo: repo}
	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, fmt.Errorf("new node: %w", err)
	}
	result, err := node.DNSResolver.LookupTXT(ctx, Peers)
	if err != nil {
		return nil, fmt.Errorf("lookup peers: %w", err)
	}
	for i := range result {
		for _, item := range strings.Split(result[i], ";") {
			log.Println("add mirror peers", item)
			peer, err := peer.AddrInfoFromString(item)
			if err != nil {
				return nil, err
			}
			node.Peering.AddPeer(*peer)
		}
	}
	return node, nil
}

// 生成节点配置
func initConfig() (*config.Config, error) {
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return nil, err
	}
	cfg.AutoNAT.ServiceMode = config.AutoNATServiceDisabled
	cfg.Routing.Type = config.NewOptionalString("dhtclient")
	cfg.Swarm.ConnMgr.GracePeriod = config.NewOptionalDuration(time.Minute)
	cfg.Swarm.ConnMgr.HighWater = config.NewOptionalInteger(40)
	cfg.Swarm.ConnMgr.LowWater = config.NewOptionalInteger(20)
	cfg.Datastore.GCPeriod = "240h"
	cfg.Bootstrap = Bootstrap
	err = fsrepo.Init(RepoPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("init config: %w", err)
	}
	err = os.WriteFile(filepath.Join(RepoPath, "swarm.key"), []byte(SwarmKey), 0644)
	if err != nil {
		return nil, fmt.Errorf("init swarm key: %w", err)
	}
	return cfg, err
}
