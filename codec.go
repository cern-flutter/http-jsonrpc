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
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
)

var (
	ErrNullResult = errors.New("Result is nil")
)

type (
	clientRequest struct {
		Version string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		Id      uint64      `json:"id"`
	}

	clientResponse struct {
		Version string           `json:"jsonrpc"`
		Result  *json.RawMessage `json:"result"`
		Error   *json.RawMessage `json:"error"`
		Id      uint64           `json:"id"`
	}

	clientError struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	codec struct {
		addr               string
		httpResponses      chan *http.Response
		lastClientResponse *clientResponse
	}
)

// New creates a new jsonrpc over HTTP client
func NewClientCodec(addr string) (rpc.ClientCodec, error) {
	return &codec{
		addr:               addr,
		httpResponses:      make(chan *http.Response),
		lastClientResponse: nil,
	}, nil
}

// WriteRequest sends the request to the server
func (c *codec) WriteRequest(request *rpc.Request, args interface{}) error {
	req := clientRequest{
		Version: "2.0",
		Method:  request.ServiceMethod,
		Params:  args,
		Id:      request.Seq,
	}
	jsonEncodedReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", c.addr, bytes.NewBuffer(jsonEncodedReq))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	c.httpResponses <- httpResp
	return nil
}

// ReadResponseHeader reads the response headers
func (c *codec) ReadResponseHeader(response *rpc.Response) error {
	httpResp := <-c.httpResponses
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		response.Error = httpResp.Status
		return nil
	}

	resp := clientResponse{}
	decoder := json.NewDecoder(httpResp.Body)
	if err := decoder.Decode(&resp); err != nil {
		return err
	}
	response.Seq = resp.Id
	if resp.Error != nil {
		error := clientError{}
		if err := json.Unmarshal(*resp.Error, &error); err != nil {
			return err
		}
		response.Error = error.Message
		return nil
	}

	c.lastClientResponse = &resp
	return nil
}

// ReadResponseBody reads the response body
func (c *codec) ReadResponseBody(reply interface{}) error {
	if reply == nil {
		return nil
	}
	if c.lastClientResponse.Result == nil {
		return ErrNullResult
	}
	return json.Unmarshal(*c.lastClientResponse.Result, reply)
}

// Closes the connection
func (c *codec) Close() error {
	return nil
}
