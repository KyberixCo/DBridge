package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxPendingRequests = 32

func validRequestID(id json.RawMessage) bool {
	if len(id) == 0 {
		return false
	}
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(id))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return false
	}
	switch value := value.(type) {
	case string:
		return true
	case json.Number:
		_, err := value.Int64()
		return err == nil
	default:
		return false
	}
}

func decodeRequest(data []byte) (JSONRPCRequest, *RPCError) {
	var req JSONRPCRequest
	if !json.Valid(data) {
		return req, &RPCError{Code: -32700, Message: "Parse error"}
	}
	if err := json.Unmarshal(data, &req); err != nil || req.JSONRPC != "2.0" || strings.TrimSpace(req.Method) == "" || (len(req.ID) > 0 && !validRequestID(req.ID)) {
		return req, &RPCError{Code: -32600, Message: "Invalid JSON-RPC request"}
	}
	return req, nil
}

func requestKey(id json.RawMessage) string {
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(id))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// RunStdio keeps diagnostics on stderr and protocol frames on stdout.
func (s *MCPServer) RunStdio(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "dbridge: MCP stdio server listening...")
	return s.runStdio(ctx, os.Stdin, os.Stdout)
}

type pendingRequest struct {
	request   JSONRPCRequest
	data      []byte
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled bool
}

// runStdio processes operations in arrival order while continuing to read
// cancellation notifications. The queue and incoming frames are bounded.
func (s *MCPServer) runStdio(ctx context.Context, input io.ReadCloser, output io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer input.Close() // unblock a reader when the parent context is cancelled
	type frame struct {
		data []byte
		err  error
	}
	frames := make(chan frame)
	go func() {
		defer close(frames)
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			data := bytes.TrimSpace(scanner.Bytes())
			if len(data) == 0 {
				continue
			}
			copy := append([]byte(nil), data...)
			select {
			case frames <- frame{data: copy}:
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case frames <- frame{err: err}:
			case <-ctx.Done():
			}
		}
	}()
	write := func(data []byte) error {
		if len(data) == 0 {
			return nil
		}
		data = append(data, '\n')
		n, err := output.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		if err != nil {
			return fmt.Errorf("could not write MCP response: %w", err)
		}
		return nil
	}
	queue := make([]*pendingRequest, 0, maxPendingRequests)
	pending := make(map[string]*pendingRequest)
	defer func() {
		for _, request := range pending {
			request.cancel()
		}
	}()
	var active *pendingRequest
	completed := make(chan []byte, 1)
	for {
		if active == nil && len(queue) > 0 {
			active, queue = queue[0], queue[1:]
			request := active
			go func() { completed <- s.ProcessJSONRPC(request.ctx, request.data, "stdio") }()
		}
		if frames == nil && active == nil && len(queue) == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case response := <-completed:
			if !active.cancelled {
				if err := write(response); err != nil {
					return err
				}
			}
			active.cancel()
			delete(pending, requestKey(active.request.ID))
			active = nil
		case incoming, ok := <-frames:
			if !ok {
				frames = nil
				continue
			}
			if incoming.err != nil {
				return fmt.Errorf("could not read MCP request: %w", incoming.err)
			}
			req, rpcErr := decodeRequest(incoming.data)
			if rpcErr != nil {
				if err := write(s.makeResponse(nil, nil, rpcErr)); err != nil {
					return err
				}
				continue
			}
			if len(req.ID) == 0 {
				if req.Method == "notifications/cancelled" {
					var params struct {
						RequestID json.RawMessage `json:"requestId"`
					}
					if json.Unmarshal(req.Params, &params) == nil && validRequestID(params.RequestID) {
						if request := pending[requestKey(params.RequestID)]; request != nil && request.request.Method != "initialize" {
							request.cancelled = true
							request.cancel()
						}
					}
				}
				continue
			}
			key := requestKey(req.ID)
			if len(pending) >= maxPendingRequests || pending[key] != nil {
				if err := write(s.makeResponse(req.ID, nil, &RPCError{Code: -32600, Message: "Request queue is full or request ID is already in progress"})); err != nil {
					return err
				}
				continue
			}
			requestCtx, requestCancel := context.WithTimeout(ctx, s.operationTimeout())
			request := &pendingRequest{request: req, data: incoming.data, ctx: requestCtx, cancel: requestCancel}
			pending[key] = request
			queue = append(queue, request)
		}
	}
}
