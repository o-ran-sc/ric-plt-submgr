/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package control

import (
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"testing"
)

func TestRmrEndpoint(t *testing.T) {

	tent := teststub.NewTestWrapper("TestRmrEndpoint")

	testEp := func(t *testing.T, val string, expect *RmrEndpoint) {
		res := NewRmrEndpoint(val)

		if expect == nil && res == nil {
			return
		}
		if res == nil {
			tent.TestError(t, "Endpoint elems for value %s expected addr %s port %d got nil", val, expect.GetAddr(), expect.GetPort())
			return
		}
		if expect.GetAddr() != res.GetAddr() || expect.GetPort() != res.GetPort() {
			tent.TestError(t, "Endpoint elems for value %s expected addr %s port %d got addr %s port %d", val, expect.GetAddr(), expect.GetPort(), res.GetAddr(), res.GetPort())
		}
		if expect.String() != res.String() {
			tent.TestError(t, "Endpoint string for value %s expected %s got %s", val, expect.String(), res.String())
		}

	}

	testEp(t, "localhost:8080", &RmrEndpoint{"localhost", 8080})
	testEp(t, "127.0.0.1:8080", &RmrEndpoint{"127.0.0.1", 8080})
	testEp(t, "localhost:70000", nil)
	testEp(t, "localhost?8080", nil)
	testEp(t, "abcdefghijklmnopqrstuvwxyz", nil)
	testEp(t, "", nil)
}

func TestRmrEndpointList(t *testing.T) {

	tent := teststub.NewTestWrapper("TestRmrEndpointList")

	epl := &RmrEndpointList{}

	// Simple add / has / delete
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8080")) == false {
		tent.TestError(t, "RmrEndpointList: 8080 add failed")
	}
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8080")) == true {
		tent.TestError(t, "RmrEndpointList: 8080 duplicate add success")
	}
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == false {
		tent.TestError(t, "RmrEndpointList: 8081 add failed")
	}
	if epl.HasEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == false {
		tent.TestError(t, "RmrEndpointList: 8081 has failed")
	}
	if epl.DelEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == false {
		tent.TestError(t, "RmrEndpointList: 8081 del failed")
	}
	if epl.HasEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == true {
		tent.TestError(t, "RmrEndpointList: 8081 has non existing success")
	}
	if epl.DelEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == true {
		tent.TestError(t, "RmrEndpointList: 8081 del non existing success")
	}
	if epl.DelEndpoint(NewRmrEndpoint("127.0.0.1:8080")) == false {
		tent.TestError(t, "RmrEndpointList: 8080 del failed")
	}

	// list delete
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8080")) == false {
		tent.TestError(t, "RmrEndpointList: 8080 add failed")
	}
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == false {
		tent.TestError(t, "RmrEndpointList: 8081 add failed")
	}
	if epl.AddEndpoint(NewRmrEndpoint("127.0.0.1:8082")) == false {
		tent.TestError(t, "RmrEndpointList: 8082 add failed")
	}

	epl2 := &RmrEndpointList{}
	if epl2.AddEndpoint(NewRmrEndpoint("127.0.0.1:9080")) == false {
		tent.TestError(t, "RmrEndpointList: othlist add 9080 failed")
	}

	if epl.DelEndpoints(epl2) == true {
		tent.TestError(t, "RmrEndpointList: delete list not existing successs")
	}

	if epl2.AddEndpoint(NewRmrEndpoint("127.0.0.1:8080")) == false {
		tent.TestError(t, "RmrEndpointList: othlist add 8080 failed")
	}
	if epl.DelEndpoints(epl2) == false {
		tent.TestError(t, "RmrEndpointList: delete list 8080,9080 failed")
	}

	if epl2.AddEndpoint(NewRmrEndpoint("127.0.0.1:8081")) == false {
		tent.TestError(t, "RmrEndpointList: othlist add 8081 failed")
	}
	if epl2.AddEndpoint(NewRmrEndpoint("127.0.0.1:8082")) == false {
		tent.TestError(t, "RmrEndpointList: othlist add 8082 failed")
	}

	if epl.DelEndpoints(epl2) == false {
		tent.TestError(t, "RmrEndpointList: delete list 8080,8081,8082,9080 failed")
	}

}
