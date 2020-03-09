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

//
// EXAMPLE HOW TO HAVE RMR STUB
//

package teststubdummy

import (
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrDummyStub struct {
	teststub.RmrStubControl
	reqMsg  int
	respMsg int
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewRmrDummyStub(desc string, rtfile string, port string, stat string, mtypeseed int) *RmrDummyStub {
	dummyStub := &RmrDummyStub{}
	dummyStub.RmrStubControl.Init(desc, rtfile, port, stat, mtypeseed)
	dummyStub.reqMsg = mtypeseed + 1
	dummyStub.respMsg = mtypeseed + 2
	return dummyStub
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *RmrDummyStub) SendReq(t *testing.T) {
	tc.Logger.Info("SendReq")
	var dummyBuf []byte = make([]byte, 100)
	params := xapptweaks.NewParams(nil)
	params.Mtype = tc.reqMsg
	params.SubId = -1
	params.Payload = dummyBuf
	params.PayloadLen = 100
	params.Meid = &xapp.RMRMeid{RanName: "TEST"}
	params.Xid = "TEST"
	params.Mbuf = nil

	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "%s", snderr.Error())
	}
	return
}

func (tc *RmrDummyStub) RecvResp(t *testing.T) bool {
	tc.Logger.Info("RecvResp")

	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != tc.respMsg {
			tc.TestError(t, "Received wrong mtype expected %d got %d, error", tc.respMsg, msg.Mtype)
			return false
		}
		return true
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return false
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func RmrDummyHandleMessage(msg *xapptweaks.RMRParams, mtypeseed int, rmr xapptweaks.XAppWrapperIf) (bool, error) {
	if msg.Mtype == mtypeseed+1 {
		var dummyBuf []byte = make([]byte, 100)
		params := xapptweaks.NewParams(nil)
		params.Mtype = mtypeseed + 2
		params.SubId = msg.SubId
		params.Payload = dummyBuf
		params.PayloadLen = 100
		params.Meid = msg.Meid
		params.Xid = msg.Xid
		params.Mbuf = nil
		rmr.GetLogger().Info("SEND DUMMY RESP: %s", params.String())
		err := rmr.RmrSend(params, 5)
		if err != nil {
			rmr.GetLogger().Error("RmrDummyHandleMessage: err(%s)", err.Error())
		}
		return true, err
	}
	return false, nil
}
