package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corehttp"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/ipfs/go-ipfs/plugin/loader"

	"github.com/coreos/go-systemd/v22/activation"
)

var (
	RepoPath = "./ipfs"
	Peers    = []string{"12D3KooWH1d6Zi8WeYbpqaP4MKv23VY6XPXMM4AoSBZq5kv6s4ey", "12D3KooWDm2o3RZsE7t2oFMqKZxYo4W1c2XwYrKbXm3qXUeVLpnp"}
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	node, err := initNode(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("inited")
	list, err := activation.Listeners()
	if err != nil {
		log.Fatal(err)
	}
	var opts []corehttp.ServeOption
	opts = append(opts, corehttp.GatewayOption(false, "/ipfs", "/ipns"))
	if len(list) > 0 {
		err = corehttp.Serve(node, list[0], opts...)
	} else {
		err = corehttp.ListenAndServe(node, "/ip4/127.0.0.1/tcp/8080", opts...)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exit")
	time.Sleep(time.Second)
}

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
		cfg, err := config.Init(os.Stdout, 2048)
		if err != nil {
			return nil, err
		}
		// 降低CPU占用
		cfg.Discovery.MDNS.Enabled = false
		cfg.Swarm.DisableBandwidthMetrics = true
		cfg.Swarm.ConnMgr.LowWater = 50
		cfg.Swarm.ConnMgr.HighWater = 100
		cfg.Swarm.EnableAutoRelay = true
		// 添加默认节点
		for i := range Peers {
			id, err := peer.Decode(Peers[i])
			if err != nil {
				return nil, err
			}
			cfg.Peering.Peers = append(cfg.Peering.Peers, peer.AddrInfo{ID: id})
		}
		err = fsrepo.Init(RepoPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("init repo: %w", err)
		}
	}
	repo, err := fsrepo.Open(RepoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTClientOption,
		Repo:    repo,
	}
	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, fmt.Errorf("new node: %w", err)
	}
	return node, nil
}
