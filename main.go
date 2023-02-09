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
	Listen    = ":8080"
	RepoPath  = "./data"
	SwarmKey  = "/key/swarm/psk/1.0.0/\n/base16/\n2508242b6ac9e665ea98eb134dd7e05497530f36876ae3ebc865f45fb104291b"
	Bootstrap = []string{
		"/dns4/bootstrap.getdeepin.org/tcp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
		"/dns6/bootstrap.getdeepin.org/tcp/4001/p2p/12D3KooWBc44S9zeb1KSdRZMCHTNpuNX3uS8SpvpbQz2SzrNsWJm",
		"/dns4/mirrors.getdeepin.org/tcp/8443/wss/p2p/12D3KooWCMcqaZQyBtRDRdVm3UsipwjG5nKR2TSk96CqRb7rNxtq",
		"/dns6/mirrors.getdeepin.org/tcp/8443/wss/p2p/12D3KooWCMcqaZQyBtRDRdVm3UsipwjG5nKR2TSk96CqRb7rNxtq",
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
	core, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, fmt.Errorf("core api:%w", err)
	}

	mirrors := "/ipns/mirrors.getdeepin.org"
	p, err := core.Name().Resolve(ctx, mirrors)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	log.Printf("resolve %s => %s\n", mirrors, p.String())
	return node, nil
}

// 生成配置
func initConfig() (*config.Config, error) {
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return nil, err
	}
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
