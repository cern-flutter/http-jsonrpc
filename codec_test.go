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
	"github.com/gorilla/mux"
	gorilla "github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"log"
	"net/http"
	"net/rpc"
	"testing"
)

var addr = "localhost:12345"

type Mock int

// Echo just echoes the input string on the output parameter
func (*Mock) Echo(r *http.Request, in *string, out *string) error {
	*out = *in
	return nil
}

// RPC server
func server() {
	var mock Mock

	server := gorilla.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	server.RegisterCodec(json.NewCodec(), "application/json-rpc")
	if err := server.RegisterService(&mock, "Mock"); err != nil {
		log.Panic(err)
	}

	router := mux.NewRouter()
	router.Handle("/rpc", server)

	// Run jsonrpc server
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Panic(err)
	}
	log.Print("Server up")
}

// init
func init() {
	go server()
}

// Test a call to echo
func TestCall(t *testing.T) {
	codec, err := NewClientCodec("http://" + addr + "/rpc")
	if err != nil {
		t.Fatal(err)
	}

	client := rpc.NewClientWithCodec(codec)

	var request string
	var reply string
	request = "Hello there"
	if err := client.Call("Mock.Echo", &request, &reply); err != nil {
		t.Fatal(err)
	}

	if request != reply {
		t.Fatal("Was expecting an echo")
	}
}

// Test a call to a missing method
func TestBadCall(t *testing.T) {
	codec, err := NewClientCodec("http://" + addr + "/rpc")
	if err != nil {
		t.Fatal(err)
	}

	client := rpc.NewClientWithCodec(codec)

	var request string
	var reply string
	request = "Hello there"
	if err := client.Call("Mock.ThisDoesNotExist", &request, &reply); err == nil {
		t.Fatal("Was expecting an error")
	}
}
