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

package xapptweaks

import (
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"sync"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrWrapper struct {
	mtx        sync.Mutex
	Rmr        *xapp.RMRClient
	CntRecvMsg uint64
	CntSentMsg uint64
}

func (tc *RmrWrapper) Lock() {
	tc.mtx.Lock()
}

func (tc *RmrWrapper) Unlock() {
	tc.mtx.Unlock()
}

func (tc *RmrWrapper) Init() {
}

func (tc *RmrWrapper) RmrSend(params *RMRParams, to time.Duration) (err error) {
	if tc.Rmr == nil {
		err = fmt.Errorf("Failed rmr object nil for %s", params.String())
		return
	}
	tc.Lock()
	status := tc.Rmr.Send(params.RMRParams, false)
	tc.Unlock()
	i := 0
	for ; i < int(to)*2 && status == false; i++ {
		tc.Lock()
		status = tc.Rmr.Send(params.RMRParams, false)
		tc.Unlock()
		if status == false {
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = fmt.Errorf("Failed with retries(%d) %s", i, params.String())
		tc.Rmr.Free(params.Mbuf)
	} else {
		tc.CntSentMsg++
	}
	return
}
