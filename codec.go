/*
 * Copyright (c) CERN 2016
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http_jsonrpc

import (
	"bytes"
	json "github.com/gorilla/rpc/v2/json2"
	"net/http"
	"net/rpc"
)

type codec struct {
	addr         string
	client       http.Client
	responses    chan *http.Response
	lastResponse *http.Response
}

// New creates a new jsonrpc over HTTP client
func NewClientCodec(addr string) (rpc.ClientCodec, error) {
	return &codec{
		addr:         addr,
		client:       http.Client{},
		responses:    make(chan *http.Response),
		lastResponse: nil,
	}, nil
}

// WriteRequest sends the request to the server
func (c *codec) WriteRequest(request *rpc.Request, args interface{}) error {
	jsonEncodedReq, err := json.EncodeClientRequest(request.ServiceMethod, args)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", c.addr, bytes.NewBuffer(jsonEncodedReq))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	c.responses <- httpResp
	return nil
}

// ReadResponseHeader reads the response headers
func (c *codec) ReadResponseHeader(response *rpc.Response) error {
	c.lastResponse = <-c.responses
	if c.lastResponse.StatusCode/100 != 2 {
		response.Error = c.lastResponse.Status
	}
	return nil
}

// ReadResponseBody reads the response body
func (c *codec) ReadResponseBody(reply interface{}) error {
	defer c.lastResponse.Body.Close()
	err := json.DecodeClientResponse(c.lastResponse.Body, reply)
	if err != nil {
		return err
	}
	return nil
}

// Closes the connection
func (c *codec) Close() error {
	return nil
}
