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
package teststub

import (
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strings"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrStubControl struct {
	RmrControl
	xapptweaks.RmrWrapper
	MsgChan  chan *xapptweaks.RMRParams
	Active   bool
	InitMsg  int
	CheckXid bool
}

func (tc *RmrStubControl) SetActive() {
	tc.Active = true
}

func (tc *RmrStubControl) IsActive() bool {
	return tc.Active
}

func (tc *RmrStubControl) SetCheckXid(val bool) {
	tc.CheckXid = val
}

func (tc *RmrStubControl) IsCheckXid() bool {
	return tc.CheckXid
}

func (tc *RmrStubControl) IsChanEmpty() bool {
	if len(tc.MsgChan) > 0 {
		return false
	}
	return true
}

func (tc *RmrStubControl) TestMsgChanEmpty(t *testing.T) {
	if tc.IsChanEmpty() == false {
		TestError(t, "(%s) message channel not empty", tc.GetDesc())
	}
}

func (tc *RmrStubControl) Init(desc string, rtfile string, port string, stat string, consumer xapp.MessageConsumer, initMsg int) {
	tc.InitMsg = initMsg
	tc.Active = false
	tc.RmrControl.Init(desc, rtfile, port)
	tc.RmrWrapper.Init()
	tc.MsgChan = make(chan *xapptweaks.RMRParams)

	tc.Rmr = xapp.NewRMRClientWithParams("tcp:"+port, 4096, 1, stat)
	tc.Rmr.SetReadyCB(tc.ReadyCB, nil)
	go tc.Rmr.Start(consumer)

	tc.WaitCB()
	allRmrStubs = append(allRmrStubs, tc)
}

func (tc *RmrStubControl) Consume(params *xapp.RMRParams) (err error) {
	defer tc.Rmr.Free(params.Mbuf)
	msg := xapptweaks.NewParams(params)

	if msg.Mtype == tc.InitMsg {
		xapp.Logger.Info("(%s) Testing message ignore %s", tc.GetDesc(), msg.String())
		tc.SetActive()
		return
	}

	if tc.IsCheckXid() == true && strings.Contains(msg.Xid, tc.GetDesc()) == false {
		xapp.Logger.Info("(%s) Ignore %s", tc.GetDesc(), msg.String())
		return
	}

	xapp.Logger.Info("(%s) Consume %s", tc.GetDesc(), msg.String())
	tc.PushMsg(msg)
	return
}

var allRmrStubs []*RmrStubControl

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func RmrStubControlWaitAlive(seconds int, mtype int, rmr xapptweaks.RmrWrapperIf) bool {

	var dummyBuf []byte = make([]byte, 100)

	params := xapptweaks.NewParams(nil)
	params.Mtype = mtype
	params.SubId = -1
	params.Payload = dummyBuf
	params.PayloadLen = 100
	params.Meid = &xapp.RMRMeid{RanName: "TESTPING"}
	params.Xid = "TESTPING"
	params.Mbuf = nil

	status := false
	i := 1
	for ; i <= seconds*2 && status == false; i++ {
		rmr.RmrSend("TESTPING", params)

		status = true
		for _, val := range allRmrStubs {
			if val.IsActive() == false {
				status = false
				break
			}
		}
		if status == true {
			break
		}
		xapp.Logger.Info("Sleep 0.5 secs and try routes again")
		time.Sleep(500 * time.Millisecond)
	}

	if status == false {
		xapp.Logger.Error("Could not initialize routes")
	}
	return status
}
