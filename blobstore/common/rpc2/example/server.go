package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/cubefs/cubefs/blobstore/cmd"
	"github.com/cubefs/cubefs/blobstore/common/config"
	"github.com/cubefs/cubefs/blobstore/common/rpc"
	"github.com/cubefs/cubefs/blobstore/common/rpc2"
	"github.com/cubefs/cubefs/blobstore/util/log"
)

type Config struct {
	cmd.Config
}

func init() {
	mod := &cmd.Module{
		Name:       "example_rpc2",
		InitConfig: initConfig,
		SetUp:      setUp,
		SetUp2:     setUp2,
		TearDown:   func() {},
	}
	cmd.RegisterModule(mod)
}

func initConfig(args []string) (*cmd.Config, error) {
	var conf Config
	config.Init("f", "", "server.conf")
	if err := config.Load(&conf); err != nil {
		return nil, err
	}
	logDir := path.Join(os.TempDir(), "example_rpc2")
	os.MkdirAll(logDir, 0o644)
	conf.AuditLog.LogDir = logDir
	conf.LogConf.Filename = path.Join(logDir, "rpc2.log")
	conf.BindAddr = listenrpc
	conf.Rpc2Server.Addresses = []rpc2.NetworkAddress{
		{Network: "tcp", Address: listenon[0]},
		{Network: "tcp", Address: listenon[1]},
	}
	return &conf.Config, nil
}

func setUp() (*rpc.Router, []rpc.ProgressHandler) {
	router := rpc.New()
	router.Handle(http.MethodGet, "/rpc", func(c *rpc.Context) { c.Respond() })
	return router, nil
}

func setUp2() (*rpc2.Router, []rpc2.Interceptor) {
	router := &rpc2.Router{}
	router.Middleware(handleMiddleware1, handleMiddleware2)
	router.Register("/ping", handlePing)
	router.Register("/kick", handleKick)
	router.Register("/error", handleError)
	router.Register("/panic", handlePanic)
	router.Register("/stream", handleStream)
	return router, []rpc2.Interceptor{interceptor{"i1"}, interceptor{"i2"}}
}

type interceptor struct{ id string }

func (i interceptor) Handle(w rpc2.ResponseWriter, req *rpc2.Request, h rpc2.Handle) error {
	log.Info("interceptor-" + i.id)
	return h(w, req)
}

func handleMiddleware1(w rpc2.ResponseWriter, req *rpc2.Request) error {
	log.Info("middleware-1")
	return nil
}

func handleMiddleware2(w rpc2.ResponseWriter, req *rpc2.Request) error {
	log.Info("middleware-2")
	return nil
}

func handleKick(_ rpc2.ResponseWriter, req *rpc2.Request) error {
	var para paraCodec
	req.ParseParameter(&para)
	return nil
}

func handleError(rpc2.ResponseWriter, *rpc2.Request) error {
	return rpc2.NewError(567, "", "")
}

func handlePanic(rpc2.ResponseWriter, *rpc2.Request) error {
	panic("handle panic")
}

func handlePing(w rpc2.ResponseWriter, req *rpc2.Request) error {
	log.Info(req.RequestHeader.String())
	var para paraCodec
	req.ParseParameter(&para)

	resp := bytes.NewBuffer(nil)
	resp.WriteString("response -> ")
	w.SetContentLength(int64(resp.Len()) + req.ContentLength)
	buff := make([]byte, req.ContentLength)
	if _, err := io.ReadFull(req.Body, buff); err != nil {
		return err
	}
	req.Body.Close()
	log.Info("body   :", string(buff))
	log.Info("trailer:", req.Trailer.M)
	w.Trailer().SetLen("server-trailer", 3)
	w.AfterBody(func() error {
		log.Info("run after body stack - 1")
		w.Trailer().Set("server-trailer", "123")
		return nil
	})
	w.AfterBody(func() error {
		log.Info("run after body stack - 2")
		return nil
	})

	w.WriteHeader(200, &para)
	w.Header().Set("ignored", "x") // ignore
	resp.Write(buff)
	_, err := w.ReadFrom(resp)
	return err
}

func handleStream(_ rpc2.ResponseWriter, req *rpc2.Request) error {
	var para paraCodec
	req.ParseParameter(&para)
	para.Value.S = "response -> " + para.Value.S

	stream := rpc2.GenericServerStream[streamReq, streamResp]{ServerStream: req.ServerStream()}
	var header, trailer rpc2.Header
	header.Set("stream-header-a", "aaa")
	trailer.Set("stream-trailer-b", "")
	stream.SetHeader(header)
	stream.SetTrailer(trailer)
	stream.SendHeader(&para)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			trailer.Set("stream-trailer-b", "bbb")
			trailer.Set("stream-trailer-x", "another")
			stream.SetTrailer(trailer)
			return nil
		}
		if err != nil {
			return err
		}
		var resp streamResp
		resp.Value = "response -> " + req.Value
		if err = stream.Send(&resp); err != nil {
			return err
		}
	}
}

func runServer() {
	cmd.Main(os.Args)
}
