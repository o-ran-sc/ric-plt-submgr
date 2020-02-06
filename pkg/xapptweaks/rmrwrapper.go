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
type RmrWrapperIf interface {
	RmrSend(desc string, params *RMRParams) (err error)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrWrapper struct {
	mtx      sync.Mutex
	Rmr      *xapp.RMRClient
	RecvChan chan *RMRParams
}

func (u *RmrWrapper) Init() {
	u.RecvChan = make(chan *RMRParams)
}

func (u *RmrWrapper) Lock() {
	u.mtx.Lock()
}

func (u *RmrWrapper) Unlock() {
	u.mtx.Unlock()
}

func (u *RmrWrapper) IsChanEmpty() bool {
	if len(u.RecvChan) > 0 {
		return false
	}
	return true
}

func (u *RmrWrapper) RmrSend(desc string, params *RMRParams) (err error) {
	if u.Rmr == nil {
		err = fmt.Errorf("(%s) RmrSend failed. Rmr object nil, %s", desc, params.String())
		return
	}
	xapp.Logger.Info("(%s) RmrSend %s", desc, params.String())
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		u.Lock()
		status = u.Rmr.Send(params.RMRParams, false)
		u.Unlock()
		if status == false {
			xapp.Logger.Info("(%s) RmrSend failed. Retry count %v, %s", desc, i, params.String())
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = fmt.Errorf("(%s) RmrSend failed. Retry count %v, %s", desc, i, params.String())
		xapp.Logger.Error("%s", err.Error())
		u.Rmr.Free(params.Mbuf)
	}
	return
}

func (u *RmrWrapper) Consume(params *xapp.RMRParams) (err error) {
	defer u.Rmr.Free(params.Mbuf)
	msg := NewParams(params)
	u.PushMsg(msg)
	return
}

func (u *RmrWrapper) PushMsg(msg *RMRParams) {
	u.RecvChan <- msg
}

func (u *RmrWrapper) WaitMsg(secs time.Duration) *RMRParams {
	if secs == 0 {
		msg := <-u.RecvChan
		return msg
	}
	select {
	case msg := <-u.RecvChan:
		return msg
	case <-time.After(secs * time.Second):
		return nil
	}
	return nil
}
