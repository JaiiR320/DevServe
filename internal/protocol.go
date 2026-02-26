package internal

import (
	"encoding/json"
	"fmt"
	"net"
)

type Request struct {
	Action string         `json:"action"`
	Args   map[string]any `json:"args,omitempty"`
}

type Response struct {
	OK    bool   `json:"ok"`
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func OkResponse(data string) *Response {
	return &Response{OK: true, Data: data}
}

func ErrResponse(err error) *Response {
	return &Response{OK: false, Error: err.Error()}
}

func SendRequest(conn net.Conn, req *Request) error {
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}
	return nil
}

func ReadRequest(conn net.Conn) (*Request, error) {
	var req Request
	err := json.NewDecoder(conn).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}
	return &req, nil
}

func SendResponse(conn net.Conn, resp *Response) error {
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}
	return nil
}

func ReadResponse(conn net.Conn) (*Response, error) {
	var resp Response
	err := json.NewDecoder(conn).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}
