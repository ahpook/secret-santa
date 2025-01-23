package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	cfgContractHash     = "contractHash"
	cfgStorageNode      = "storage_node"
	cfgStorageContainer = "storage_container"
	cfgListenAddress    = "listen_address"
)

type Server struct {
	p            *pool.Pool
	acc          *wallet.Account
	act          *actor.Actor
	gasAct       *nep17.Token
	contractHash util.Uint160
	cnrID        cid.ID
	log          *zap.Logger
	rpcCli       *rpcclient.Client
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		fmt.Println("\nShutting down gracefully...")

		s.cleanup()
		die(ctx.Err())

		close(done)
	}()

	err = s.Listen(ctx)
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
		s.cleanup()
	}

	<-done
	fmt.Println("Shutdown complete.")
}

func NewServer(ctx context.Context) (*Server, error) {
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

	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &Server{
		p:            p,
		acc:          acc,
		act:          act,
		rpcCli:       rpcCli,
		contractHash: contractHash,
		gasAct:       nep17.New(act, gas.Hash),
		cnrID:        cnrID,
		log:          log,
	}, nil
}

func (s *Server) Listen(ctx context.Context) error {
	http.DefaultServeMux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("PUT request received")
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to parse file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		var buf bytes.Buffer
		_, err = buf.ReadFrom(file)
		if err != nil {
			http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fileContent := buf.Bytes()

		filename := r.FormValue("filename")

		err = s.uploadFileToFrostFS(ctx, fileContent, filename)
		if err != nil {
			http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "File uploaded successfully")
	})

	http.DefaultServeMux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("GET request received")
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Missing 'filename' parameter", http.StatusBadRequest)
			return
		}

		fileContent, err := s.getFileFromFrostFS(ctx, filename)
		if err != nil {
			http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(fileContent)
	})

	http.DefaultServeMux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("DELETE request received")
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Missing 'filename' parameter", http.StatusBadRequest)
			return
		}

		err := s.deleteFileFromFrostFS(ctx, filename)
		if err != nil {
			http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "File deleted successfully")
	})

	s.log.Info("HTTP server started")
	return http.ListenAndServe(viper.GetString(cfgListenAddress), nil)
}

func (s *Server) getFileFromFrostFS(ctx context.Context, filename string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) deleteFileFromFrostFS(ctx context.Context, filename string) error {
	return fmt.Errorf("not implemented")
}

func (s *Server) uploadFileToFrostFS(ctx context.Context, fileContent []byte, filename string) error {
	var ownerID user.ID
	user.IDFromKey(&ownerID, s.acc.PrivateKey().PrivateKey.PublicKey)

	obj := object.New()
	obj.SetContainerID(s.cnrID)
	obj.SetOwnerID(ownerID)

	attr := *object.NewAttribute()
	attr.SetKey("filename")
	attr.SetValue(filename)

	hash := sha256.Sum256(fileContent)

	hashString := hex.EncodeToString(hash[:])

	fileHash := *object.NewAttribute()
	fileHash.SetKey("filehash")
	fileHash.SetValue(hashString)

	obj.SetAttributes(attr, fileHash)

	var prm pool.PrmObjectPut
	prm.SetPayload(bytes.NewReader(fileContent))
	prm.SetHeader(*obj)

	objID, err := s.p.PutObject(ctx, prm)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	fmt.Println(objID.ObjectID)

	frostFSAddr := s.cnrID.EncodeToString() + "/" + objID.ObjectID.EncodeToString()
	s.log.Info("Object uploaded to FrostFS", zap.String("address", frostFSAddr))

	ownerIdHash, _ := ownerID.ScriptHash()

	result, err := s.act.Call(
		s.contractHash,
		"addDocument",
		ownerIdHash,
		filename,
		fileContent,
	)
	if err != nil {
		return fmt.Errorf("invoke AddDocument: %w", err)
	}

	fmt.Println(result)

	s.log.Info("Smart contract invoked to register document",
				zap.String("filename", filename),
				zap.String("address", frostFSAddr),)

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

func (s *Server) cleanup() {
	if s.rpcCli != nil {
		s.rpcCli.Close()
	}

	if s.p != nil {
		s.p.Close()
	}

	if s.log != nil {
		_ = s.log.Sync()
	}

	fmt.Println("Resources released successfully.")
}
