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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"net/http"
	"sync"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type httpEventWaiter struct {
	resultChan   chan bool
	nextActionOk bool
}

func (msg *httpEventWaiter) SetResult(res bool) {
	msg.resultChan <- res
}

func (msg *httpEventWaiter) WaitResult(t *testing.T) bool {
	select {
	case result := <-msg.resultChan:
		return result
	case <-time.After(15 * time.Second):
		testError(t, "Waiter not received result status from case within 15 secs")
		return false
	}
	testError(t, "Waiter error in default branch")
	return false
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingHttpRtmgrStub struct {
	sync.Mutex
	desc        string
	port        string
	eventWaiter *httpEventWaiter
}

func (hc *testingHttpRtmgrStub) NextEvent(eventWaiter *httpEventWaiter) {
	hc.Lock()
	defer hc.Unlock()
	hc.eventWaiter = eventWaiter
}

func (hc *testingHttpRtmgrStub) AllocNextEvent(nextAction bool) *httpEventWaiter {
	eventWaiter := &httpEventWaiter{
		resultChan:   make(chan bool),
		nextActionOk: nextAction,
	}
	hc.NextEvent(eventWaiter)
	return eventWaiter
}

func (hc *testingHttpRtmgrStub) http_handler(w http.ResponseWriter, r *http.Request) {

	hc.Lock()
	defer hc.Unlock()

	var req rtmgr_models.XappSubscriptionData
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
	}
	xapp.Logger.Info("(%s) handling Address=%s Port=%d SubscriptionID=%d", hc.desc, *req.Address, *req.Port, *req.SubscriptionID)

	var code int = 0
	switch r.Method {
	case http.MethodPost:
		code = 201
		if hc.eventWaiter != nil {
			if hc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	case http.MethodDelete:
		code = 200
		if hc.eventWaiter != nil {
			if hc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	default:
		code = 200
	}

	waiter := hc.eventWaiter
	hc.eventWaiter = nil
	if waiter != nil {
		waiter.SetResult(true)
	}
	xapp.Logger.Info("(%s) Method=%s Reply with code %d", hc.desc, r.Method, code)
	w.WriteHeader(code)

}

func (hc *testingHttpRtmgrStub) run() {
	http.HandleFunc("/", hc.http_handler)
	http.ListenAndServe("localhost:"+hc.port, nil)
}

func createNewHttpRtmgrStub(desc string, port string) *testingHttpRtmgrStub {
	hc := &testingHttpRtmgrStub{}
	hc.desc = desc
	hc.port = port
	return hc
}
