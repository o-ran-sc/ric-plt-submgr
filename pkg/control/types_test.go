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
	"testing"
)

func TestRmrEndpoint(t *testing.T) {

	testEp := func(t *testing.T, val string, expect *RmrEndpoint) {
		res := NewRmrEndpoint(val)

		if expect == nil && res == nil {
			return
		}
		if res == nil {
			testError(t, "Endpoint elems for value %s expected addr %s port %d got nil", val, expect.GetAddr(), expect.GetPort())
			return
		}
		if expect.GetAddr() != res.GetAddr() || expect.GetPort() != res.GetPort() {
			testError(t, "Endpoint elems for value %s expected addr %s port %d got addr %s port %d", val, expect.GetAddr(), expect.GetPort(), res.GetAddr(), res.GetPort())
		}
		if expect.String() != res.String() {
			testError(t, "Endpoint string for value %s expected %s got %s", val, expect.String(), res.Get())
		}

	}

	testEp(t, "localhost:8080", &RmrEndpoint{"localhost", 8080})
	testEp(t, "127.0.0.1:8080", &RmrEndpoint{"127.0.0.1", 8080})
	testEp(t, "localhost:70000", nil)
	testEp(t, "localhost?8080", nil)
	testEp(t, "abcdefghijklmnopqrstuvwxyz", nil)
	testEp(t, "", nil)
}

func TestAction(t *testing.T) {

	testActionString := func(t *testing.T, val int, str string) {
		if Action(val).String() != str {
			testError(t, "String for value %d expected %s got %s", val, str, Action(val).String())
		}
	}

	testActionString(t, 0, "CREATE")
	testActionString(t, 1, "UPDATE")
	testActionString(t, 2, "NONE")
	testActionString(t, 3, "DELETE")
	testActionString(t, 5, "UNKNOWN")
	testActionString(t, 6, "UNKNOWN")
	testActionString(t, 7, "UNKNOWN")
	testActionString(t, 10, "UNKNOWN")
}
