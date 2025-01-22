package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	morphclient "git.frostfs.info/TrueCloudLab/frostfs-node/pkg/morph/client"
	"git.frostfs.info/TrueCloudLab/frostfs-node/pkg/morph/subscriber"
	"git.frostfs.info/TrueCloudLab/frostfs-node/pkg/util/logger"
	cid "git.frostfs.info/TrueCloudLab/frostfs-sdk-go/container/id"
	"git.frostfs.info/TrueCloudLab/frostfs-sdk-go/object"
	"git.frostfs.info/TrueCloudLab/frostfs-sdk-go/pool"
	"git.frostfs.info/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/actor"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	cfgRPCEndpoint      = "rpc_endpoint"
	cfgRPCEndpointWS    = "rpc_endpoint_ws"
	cfgWallet           = "wallet"
	cfgPassword         = "password"
	cfgContractHash     = "goofyahhdocuments_contract"
	cfgStorageNode      = "storage_node"
	cfgStorageContainer = "storage_container"
	cfgListenAddress    = "listen_address"
)

type Server struct {
	p        *pool.Pool
	acc      *wallet.Account
	act      *actor.Actor
	gasAct   *nep17.Token
	nyanHash util.Uint160
	cnrID    cid.ID
	log      *zap.Logger
	rpcCli   *rpcclient.Client
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if len(os.Args) != 2 {
		die(fmt.Errorf("invalid args: %v", os.Args))
	}

	viper.GetViper().SetConfigType("yml")

	f, err := os.Open(os.Args[1])
	die(err)
	die(viper.GetViper().ReadConfig(f))
	die(f.Close())

	s, err := NewServer(ctx)
	die(err)

	die(s.Listen(ctx))
}

func ExtractAccountFromWallet(ctx context.Context) (*Server, error) {
	rpcCli, err := rpcclient.New(ctx, viper.GetString(cfgRPCEndpoint), rpcclient.Options{})
	if err != nil {
		return nil, err
	}

	w, err := wallet.NewWalletFromFile(viper.GetString(cfgWallet))
	if err != nil {
		return nil, err
	}

	acc := w.GetAccount(w.GetChangeAddress())
	if err = acc.Decrypt(viper.GetString(cfgPassword), w.Scrypt); err != nil {
		return nil, err
	}

	act, err := actor.NewSimple(rpcCli, acc)
	if err != nil {
		return nil, err
	}

	p, err := createPool(ctx, acc, viper.GetString(cfgStorageNode))
	if err != nil {
		return nil, err
	}

	contractHash, err := util.Uint160DecodeStringLE(viper.GetString(cfgContractHash))
	if err != nil {
		return nil, err
	}

	var cnrID cid.ID
	if err = cnrID.DecodeString(viper.GetString(cfgStorageContainer)); err != nil {
		return nil, err
	}

	neoClient, err := morphclient.New(ctx, acc.PrivateKey(),
		morphclient.WithEndpoints(morphclient.Endpoint{Address: viper.GetString(cfgRPCEndpointWS), Priority: 1}),
	)
	if err != nil {
		return nil, fmt.Errorf("new morph client: %w", err)
	}

	// if err = neoClient.EnableNotarySupport(); err != nil {
	// 	return nil, err
	// }

	params := new(subscriber.Params)
	params.Client = neoClient
	l, err := logger.NewLogger(nil)
	if err != nil {
		return nil, err
	}
	params.Log = l
	_, err = subscriber.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// if err = sub.SubscribeForNotaryRequests(acc.ScriptHash()); err != nil {
	// 	return nil, err
	// }

	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &Server{
		p:        p,
		acc:      acc,
		act:      act,
		rpcCli:   rpcCli,
		nyanHash: contractHash,
		gasAct:   nep17.New(act, gas.Hash),
		cnrID:    cnrID,
		log:      log,
	}, nil
}

func (s *Server) Listen(ctx context.Context) error {
	// Load the local file
	filePath := "goofyahhdocument.txt" // Path to the file in the root of your repo
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Upload the file to FrostFS
	err = s.uploadFileToFrostFS(ctx, fileContent, "goofyahhdocument.txt")
	if err != nil {
		return fmt.Errorf("failed to upload file to FrostFS: %w", err)
	}

	// s.log.Info("File uploaded successfully", zap.String("objectID", objID.String()))

	// Start the HTTP server (if needed)
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("upload request")

		// Handle file upload via HTTP (if needed)
	})

	return http.ListenAndServe(viper.GetString(cfgListenAddress), nil)
}

func (s *Server) uploadFileToFrostFS(ctx context.Context, fileContent []byte, filename string) error {
	var ownerID user.ID
	user.IDFromKey(&ownerID, s.acc.PrivateKey().PrivateKey.PublicKey)

	// Create a new object
	obj := object.New()
	obj.SetContainerID(s.cnrID)
	obj.SetOwnerID(ownerID)

	attr := *object.NewAttribute()
	attr.SetKey("filename")
	attr.SetValue(filename)

	// Add an attribute for the filename
	obj.SetAttributes(attr)

	// Prepare the object for upload
	var prm pool.PrmObjectPut
	prm.SetPayload(bytes.NewReader(fileContent)) // The file bytes as payload
	prm.SetHeader(*obj)                          // Set object header (container ID, owner, attributes)

	// Upload the object to FrostFS
	objID, err := s.p.PutObject(ctx, prm)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	fmt.Print(objID.ObjectID)

	return nil
}

func createPool(ctx context.Context, acc *wallet.Account, addr string) (*pool.Pool, error) {
	var prm pool.InitParameters
	prm.SetKey(&acc.PrivateKey().PrivateKey)
	prm.AddNode(pool.NewNodeParam(1, addr, 1))
	prm.SetNodeDialTimeout(5 * time.Second)

	p, err := pool.NewPool(prm)
	if err != nil {
		return nil, fmt.Errorf("new Pool: %w", err)
	}

	err = p.Dial(ctx)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return p, nil
}

func die(err error) {
	if err == nil {
		return
	}

	debug.PrintStack()
	_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
