// Copyright 2024 The CubeFS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package rpc2

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cubefs/cubefs/blobstore/common/rpc"
	auth_proto "github.com/cubefs/cubefs/blobstore/common/rpc/auth/proto"
	"github.com/cubefs/cubefs/blobstore/util"
	"github.com/cubefs/cubefs/blobstore/util/defaulter"
	"github.com/cubefs/cubefs/blobstore/util/retry"
)

type Client struct {
	Connector       Connector       `json:"-"`
	ConnectorConfig ConnectorConfig `json:"connector"`

	Retry   int              `json:"retry"`
	RetryOn func(error) bool `json:"-"`
	// | Request | Response Header |   Response Body  |
	// |      Request Timeout      | Response Timeout |
	// |                 Timeout                      |
	Timeout         util.Duration `json:"timeout"`
	RequestTimeout  util.Duration `json:"request_timeout"`
	ResponseTimeout util.Duration `json:"response_timeout"`

	Auth auth_proto.Config `json:"auth"`

	Selector rpc.Selector `json:"-"` // lb client
	LbConfig struct {
		Hosts              []string `json:"hosts"`
		BackupHosts        []string `json:"backup_hosts"`
		HostTryTimes       int      `json:"host_try_times"`
		FailRetryIntervalS int      `json:"fail_retry_interval_s"`
		MaxFailsPeriodS    int      `json:"max_fails_period_s"`
	} `json:"lb"`

	// dead-lock copied Client when initOnce == 1
	initOnce uint32 // 0 uninitialised, 1 doing, 2 done
}

// Request simple request, parameter and result both in body.
func (c *Client) Request(ctx context.Context, addr, path string,
	para Marshaler, ret Unmarshaler,
) (err error) {
	req, err := NewRequest(ctx, addr, path, nil, Codec2Reader(para))
	if err != nil {
		return err
	}
	err = c.DoWith(req, ret)
	req.reuse()
	return
}

func (c *Client) DoWith(req *Request, ret Unmarshaler) error {
	resp, err := c.Do(req, ret)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *Client) Do(req *Request, ret Unmarshaler) (resp *Response, err error) {
	if c.lockInit() {
		defaulter.LessOrEqual(&c.Retry, 3)
		c.newSelector()
		if c.Connector == nil {
			c.Connector = defaultConnector(c.ConnectorConfig)
		}
		if c.RetryOn == nil {
			c.RetryOn = func(err error) bool { return DetectStatusCode(err) >= 500 }
		}
		atomic.StoreUint32(&c.initOnce, 2)
	}

	var lbHost rpc.UniqueHost
	var lbHosts []rpc.UniqueHost
	useLb := req.RemoteAddr == ""
	if useLb && c.Selector == nil {
		return nil, ErrConnNoAddress
	}

	if c.Auth.EnableAuth && c.Auth.Secret != "" {
		req.Header.Set(auth_proto.TokenHeaderKey, auth_proto.Encode(time.Now().Unix(),
			[]byte(req.RemotePath), []byte(c.Auth.Secret)))
	}
	if req.Header.Get(rpc.HeaderUA) == "" {
		req.Header.Set(rpc.HeaderUA, rpc.UserAgent)
	}
	for _, opt := range req.opts {
		opt(req)
	}
	err = retry.Timed(c.Retry, 1).RuptOn(func() (bool, error) {
		if useLb {
			if len(lbHosts) == 0 {
				if lbHosts = c.Selector.GetAvailableHosts(); len(lbHosts) == 0 {
					return true, ErrConnNoAddress
				}
			}
			lbHost = lbHosts[0]
			lbHosts = lbHosts[1:]
			req.RemoteAddr = lbHost.Host()
		}

		resp, err = c.do(req, ret)
		if err != nil {
			if c.RetryOn != nil && !c.RetryOn(err) {
				return true, err
			}
			if req.Body == nil || req.GetBody == nil {
				return true, err
			}
			span := req.Span()
			body, errBody := req.GetBody()
			if errBody != nil {
				span.Info("retry to get body ->", errBody)
				return true, err
			}
			req.Body = clientNopBody(body)
			if useLb {
				span.Debug("retry to set fail lb host ->", lbHost.ID(), lbHost.Host())
				c.Selector.SetFailHost(lbHost)
			}
			span.Info("retry to next ->", err)
			return false, err
		}
		return true, nil
	})
	return
}

func (c *Client) Close() error {
	if c.Selector != nil {
		c.Selector.Close()
	}
	if c.Connector == nil {
		return nil
	}
	return c.Connector.Close()
}

func (c *Client) lockInit() bool {
	if atomic.LoadUint32(&c.initOnce) >= 2 {
		return false
	}
	for !atomic.CompareAndSwapUint32(&c.initOnce, 0, 1) {
		if atomic.LoadUint32(&c.initOnce) >= 2 {
			return false
		}
	}
	return true
}

func (c *Client) do(req *Request, ret Unmarshaler) (*Response, error) {
	req.Header.SetStable()
	req.Trailer.SetStable()

	span := req.Span().WithOperation("client.do")

	conn, err := c.Connector.Get(req.Context(), req.RemoteAddr)
	if err != nil {
		span.Warn("get connection ->", err)
		return nil, err
	}
	req.client = c
	req.conn = conn
	span.Debugf("get connection -> stream(%d, %v, %v)",
		conn.ID(), conn.LocalAddr(), conn.RemoteAddr())

	resp, err := req.request(c.requestDeadline(req.Context()))
	if err != nil {
		span.Warn("send request ->", err)
		c.Connector.Put(req.Context(), req.conn, true)
		return nil, err
	}
	if err = resp.ParseResult(ret); err != nil {
		span.Warn("parse result ->", err)
		resp.Body.Close()
		return nil, err
	}
	req.conn.SetReadDeadline(c.responseDeadline(req.Context()))
	return resp, nil
}

func (c *Client) requestDeadline(ctx context.Context) time.Time {
	var timeout, reqTimeout time.Time
	if c.Timeout.Duration > 0 {
		timeout = time.Now().Add(c.Timeout.Duration)
	}
	if c.RequestTimeout.Duration > 0 {
		reqTimeout = time.Now().Add(c.RequestTimeout.Duration)
	}
	return beforeContextDeadline(ctx, latestTime(timeout, reqTimeout))
}

func (c *Client) responseDeadline(ctx context.Context) time.Time {
	var timeout, respTimeout time.Time
	if c.Timeout.Duration > 0 {
		timeout = time.Now().Add(c.Timeout.Duration)
	}
	if c.ResponseTimeout.Duration > 0 {
		respTimeout = time.Now().Add(c.ResponseTimeout.Duration)
	}
	return beforeContextDeadline(ctx, latestTime(timeout, respTimeout))
}

func (c *Client) newSelector() {
	if c.Selector != nil {
		return
	}
	lb := &c.LbConfig
	if hosts := len(lb.Hosts) + len(lb.BackupHosts); hosts > 0 {
		defaulter.LessOrEqual(&lb.HostTryTimes, hosts)
		defaulter.Equal(&lb.MaxFailsPeriodS, 10)
		defaulter.Equal(&lb.FailRetryIntervalS, 300)
		c.Selector = rpc.NewSelector(&rpc.LbConfig{
			Hosts:              lb.Hosts[:],
			BackupHosts:        lb.BackupHosts[:],
			HostTryTimes:       lb.HostTryTimes,
			FailRetryIntervalS: lb.FailRetryIntervalS,
			MaxFailsPeriodS:    lb.MaxFailsPeriodS,
		})
	}
}

func NewRequest(ctx context.Context, addr, path string, para Marshaler, body io.Reader) (*Request, error) {
	ctx = ContextWithTrace(ctx)
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = io.NopCloser(body)
	}
	if para == nil {
		para = NoParameter
	}

	req := getRequest()
	req.RemotePath = path
	req.TraceID = getSpan(ctx).TraceID()
	if psize := para.Size(); psize > 0 {
		if cap(req.Parameter) >= psize {
			nn, err := para.MarshalTo(req.Parameter[:psize])
			if err != nil {
				return nil, err
			}
			if nn != psize {
				return nil, io.ErrShortWrite
			}
			req.Parameter = req.Parameter[:psize]
		} else {
			paraData, err := para.Marshal()
			if err != nil {
				return nil, err
			}
			req.Parameter = paraData
		}
	}
	req.RemoteAddr = addr
	req.ctx = ctx
	req.Body = clientNopBody(rc)
	req.AfterBody = func() error { return nil }

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return io.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		case *codecReadWriter:
			req.ContentLength = int64(v.Size())
			marshaler := v.marshaler
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(Codec2Reader(marshaler)), nil
			}
		default:
		}
		if req.ContentLength == 0 {
			if sized, ok := body.(interface{ Size() int }); ok {
				req.ContentLength = int64(sized.Size())
			}
		}
		if req.GetBody != nil && req.ContentLength == 0 {
			req.Body = NoBody
			req.GetBody = func() (io.ReadCloser, error) { return NoBody, nil }
		}
	}
	return req, nil
}

type StreamClient[Req any, Res any] struct {
	Client *Client
}

func (sc *StreamClient[Req, Res]) Streaming(req *Request, ret Unmarshaler) (StreamingClient[Req, Res], error) {
	resp, err := sc.Client.Do(req, ret)
	if err != nil {
		return nil, err
	}
	cs := &clientStream{
		req:     req,
		header:  resp.Header,
		trailer: resp.Trailer.ToHeader(),
	}
	return &GenericClientStream[Req, Res]{ClientStream: cs}, nil
}

func NewStreamRequest(ctx context.Context, addr, path string, para Marshaler) (*Request, error) {
	if para == nil {
		para = NoParameter
	}
	req, err := NewRequest(ctx, addr, path, nil, Codec2Reader(para))
	if err != nil {
		return nil, err
	}
	req.StreamCmd = StreamCmd_SYN
	req.ContentLength = int64(para.Size())
	return req, nil
}
