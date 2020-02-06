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
	"encoding/json"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"net/http"
	"sync"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type HttpEventWaiter struct {
	resultChan   chan bool
	nextActionOk bool
}

func (msg *HttpEventWaiter) SetResult(res bool) {
	msg.resultChan <- res
}

func (msg *HttpEventWaiter) WaitResult(t *testing.T) bool {
	select {
	case result := <-msg.resultChan:
		return result
	case <-time.After(15 * time.Second):
		teststub.TestError(t, "Waiter not received result status from case within 15 secs")
		return false
	}
	teststub.TestError(t, "Waiter error in default branch")
	return false
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingHttpRtmgrStub struct {
	sync.Mutex
	desc        string
	port        string
	eventWaiter *HttpEventWaiter
}

func (tc *testingHttpRtmgrStub) NextEvent(eventWaiter *HttpEventWaiter) {
	tc.Lock()
	defer tc.Unlock()
	tc.eventWaiter = eventWaiter
}

func (tc *testingHttpRtmgrStub) AllocNextEvent(nextAction bool) *HttpEventWaiter {
	eventWaiter := &HttpEventWaiter{
		resultChan:   make(chan bool),
		nextActionOk: nextAction,
	}
	tc.NextEvent(eventWaiter)
	return eventWaiter
}

func (tc *testingHttpRtmgrStub) http_handler(w http.ResponseWriter, r *http.Request) {

	tc.Lock()
	defer tc.Unlock()

	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		var req rtmgr_models.XappSubscriptionData
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
		}
		xapp.Logger.Info("(%s) handling SubscriptionID=%d Address=%s Port=%d", tc.desc, *req.SubscriptionID, *req.Address, *req.Port)
	}
	if r.Method == http.MethodPut {
		var req rtmgr_models.XappList
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
		}
		xapp.Logger.Info("(%s) handling put", tc.desc)
	}

	var code int = 0
	switch r.Method {
	case http.MethodPost:
		code = 201
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	case http.MethodDelete:
		code = 200
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	case http.MethodPut:
		code = 201
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	default:
		code = 200
	}

	waiter := tc.eventWaiter
	tc.eventWaiter = nil
	if waiter != nil {
		waiter.SetResult(true)
	}
	xapp.Logger.Info("(%s) Method=%s Reply with code %d", tc.desc, r.Method, code)
	w.WriteHeader(code)

}

func (tc *testingHttpRtmgrStub) run() {
	http.HandleFunc("/", tc.http_handler)
	http.ListenAndServe("localhost:"+tc.port, nil)
}

func createNewHttpRtmgrStub(desc string, port string) *testingHttpRtmgrStub {
	tc := &testingHttpRtmgrStub{}
	tc.desc = desc
	tc.port = port
	return tc
}
