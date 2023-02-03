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
	"syscall"
	"time"

	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/corehttp"
	"github.com/ipfs/kubo/repo/fsrepo"

	"github.com/ipfs/kubo/plugin/loader"

	"github.com/coreos/go-systemd/v22/activation"
)

var (
	Listen   = ":8080"
	RepoPath = "./data"
	Peers    = []string{"12D3KooWH1d6Zi8WeYbpqaP4MKv23VY6XPXMM4AoSBZq5kv6s4ey"}
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
	}
	var opts []corehttp.ServeOption
	opts = append(opts, corehttp.GatewayOption(false, "/ipfs", "/ipns"))
	err = corehttp.Serve(node, l, opts...)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exiting")
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
	core, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, fmt.Errorf("core api:%w", err)
	}
	_ = core
	// p, err := core.Name().Resolve(ctx, "/ipns/mirrors.myml.dev/deepin")
	// if err != nil {
	// 	return nil, fmt.Errorf("resolve path: %w", err)
	// }
	// log.Println("resolve /ipns/mirrors.myml.dev/deepin", "=>", p.String())
	return node, nil
}

var swarmKey = `/key/swarm/psk/1.0.0/
/base16/
2508242b6ac9e665ea98eb134dd7e05497530f36876ae3ebc865f45fb104291b
`

func initConfig() (*config.Config, error) {
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return nil, err
	}
	// 降低CPU占用
	cfg.Discovery.MDNS.Enabled = false
	cfg.Swarm.DisableBandwidthMetrics = true
	cfg.Swarm.RelayClient.Enabled = config.True
	cfg.Bootstrap = []string{"/ip6/2a00:b700::2:331/tcp/4001/p2p/12D3KooWK2aPBA7XMMtFeGmBL16J9d3fQ1RUm4Wirdmpg69tgURj"}

	err = fsrepo.Init(RepoPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("init config: %w", err)
	}
	err = os.WriteFile(filepath.Join(RepoPath, "swarm.key"), []byte(swarmKey), 0644)
	if err != nil {
		return nil, fmt.Errorf("init swarm key: %w", err)
	}
	return cfg, err
}
