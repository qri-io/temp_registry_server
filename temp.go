package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/qri-io/qri/config"
	"github.com/qri-io/qri/dsref"
	"github.com/qri-io/qri/lib"
	"github.com/qri-io/qri/registry"
	"github.com/qri-io/qri/registry/regserver"
	"github.com/qri-io/qri/remote"
	"github.com/qri-io/qri/repo/gen"
)

// NewTempRepoRegistry creates a temporary repo & builds a registry atop it.
// callers should always call the returned cleanup function when finished to
// remove temp files
func NewTempRepoRegistry(ctx context.Context) (*lib.Instance, registry.Registry, func(), error) {
	RootPath, err := ioutil.TempDir("", "temp_repo_registry")
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}
	log.Infof("temp registry location: %s", RootPath)

	// Create directory for new Qri repo.
	QriPath := filepath.Join(RootPath, "qri")
	err = os.MkdirAll(QriPath, os.ModePerm)
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}

	// Create directory for new IPFS repo.
	IPFSPath := filepath.Join(QriPath, "ipfs")
	err = os.MkdirAll(IPFSPath, os.ModePerm)
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}

	g := gen.NewCryptoSource()

	cfg := config.DefaultConfig()
	cfgPath := filepath.Join(QriPath, "config.yaml")

	cfg.SetPath(cfgPath)
	cfg.Profile = config.DefaultProfile()

	if cfg.P2P.PrivKey == "" {
		privKey, peerID := g.GeneratePrivateKeyAndPeerID()
		cfg.P2P.PrivKey = privKey
		cfg.P2P.PeerID = peerID
	}
	if cfg.Profile.PrivKey == "" {
		cfg.Profile.PrivKey = cfg.P2P.PrivKey
		cfg.Profile.ID = cfg.P2P.PeerID
		cfg.Profile.Peername = g.GenerateNickname(cfg.P2P.PeerID)
	}

	cfg.API.Address = "/ip4/127.0.0.1/tcp/99999"
	cfg.RPC.Enabled = false
	cfg.Registry.Location = ""
	cfg.Remote = &config.Remote{
		Enabled:          true,
		AcceptSizeMax:    -1,
		AcceptTimeoutMs:  -1,
		RequireAllBlocks: false,
		AllowRemoves:     true,
	}

	err = lib.Setup(lib.SetupParams{
		RepoPath:  QriPath,
		SetupIPFS: true,
		Config:    cfg,
		Generator: g,
	})
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}

	cleanup := func() {
		os.RemoveAll(RootPath)
	}

	opts := []lib.Option{
		lib.OptSetIPFSPath(IPFSPath),
	}

	inst, err := lib.NewInstance(ctx, QriPath, opts...)
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}
	localResolver := dsref.SequentialResolver(
		inst.Dscache(),
		inst.Repo(),
	)

	rem, err := remote.NewRemote(inst.Node(), cfg.Remote, localResolver)
	if err != nil {
		return nil, registry.Registry{}, nil, err
	}

	reg := registry.Registry{
		Remote:   rem,
		Profiles: registry.NewMemProfiles(),
		Search:   regserver.MockRepoSearch{Repo: inst.Repo()},
	}

	return inst, reg, cleanup, nil
}
