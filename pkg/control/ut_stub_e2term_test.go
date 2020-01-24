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
  "gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
  "gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
  "gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
  "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
  "testing"
  "time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var e2t_e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingE2termStub struct {
  testingRmrStubControl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func createNewE2termStub(desc string, rtfile string, port string, stat string) *testingE2termStub {
  e2termCtrl := &testingE2termStub{}
  e2termCtrl.testingRmrStubControl.init(desc, rtfile, port, stat, e2termCtrl)
  return e2termCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *testingE2termStub) Consume(params *xapp.RMRParams) (err error) {
  xapp.Rmr.Free(params.Mbuf)
  params.Mbuf = nil
  msg := &RMRParams{params}

  if params.Mtype == 55555 {
    xapp.Logger.Info("(%s) Testing message ignore %s", tc.desc, msg.String())
    tc.active = true
    return
  }

  xapp.Logger.Info("(%s) Consume %s", tc.desc, msg.String())
  tc.rmrConChan <- msg
  return
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e2termConn *testingE2termStub) handle_e2term_subs_req(t *testing.T) (*e2ap.E2APSubscriptionRequest, *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_subs_req", e2termConn.desc)
  e2SubsReq := e2t_e2asnpacker.NewPackerSubscriptionRequest()

  //---------------------------------
  // e2term activity: Recv Subs Req
  //---------------------------------
  select {
  case msg := <-e2termConn.rmrConChan:
    e2termConn.DecMsgCnt()
    if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_REQ"] {
      testError(t, "(%s) Received wrong mtype expected %s got %s, error", e2termConn.desc, "RIC_SUB_REQ", xapp.RicMessageTypeToName[msg.Mtype])
    } else {
      xapp.Logger.Info("(%s) Recv Subs Req", e2termConn.desc)
      packedData := &packer.PackedData{}
      packedData.Buf = msg.Payload
      unpackerr := e2SubsReq.UnPack(packedData)
      if unpackerr != nil {
        testError(t, "(%s) RIC_SUB_REQ unpack failed err: %s", e2termConn.desc, unpackerr.Error())
      }
      geterr, req := e2SubsReq.Get()
      if geterr != nil {
        testError(t, "(%s) RIC_SUB_REQ get failed err: %s", e2termConn.desc, geterr.Error())
      }
      return req, msg
    }
  case <-time.After(15 * time.Second):
    testError(t, "(%s) Not Received RIC_SUB_REQ within 15 secs", e2termConn.desc)
  }
  return nil, nil
}

func (e2termConn *testingE2termStub) handle_e2term_subs_resp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_subs_resp", e2termConn.desc)
  e2SubsResp := e2t_e2asnpacker.NewPackerSubscriptionResponse()

  //---------------------------------
  // e2term activity: Send Subs Resp
  //---------------------------------
  xapp.Logger.Info("(%s) Send Subs Resp", e2termConn.desc)

  resp := &e2ap.E2APSubscriptionResponse{}

  resp.RequestId.Id = req.RequestId.Id
  resp.RequestId.Seq = req.RequestId.Seq
  resp.FunctionId = req.FunctionId

  resp.ActionAdmittedList.Items = make([]e2ap.ActionAdmittedItem, len(req.ActionSetups))
  for index := int(0); index < len(req.ActionSetups); index++ {
    resp.ActionAdmittedList.Items[index].ActionId = req.ActionSetups[index].ActionId
  }

  for index := uint64(0); index < 1; index++ {
    item := e2ap.ActionNotAdmittedItem{}
    item.ActionId = index
    item.Cause.Content = 1
    item.Cause.CauseVal = 1
    resp.ActionNotAdmittedList.Items = append(resp.ActionNotAdmittedList.Items, item)
  }

  e2SubsResp.Set(resp)
  xapp.Logger.Debug("%s", e2SubsResp.String())
  packerr, packedMsg := e2SubsResp.Pack(nil)
  if packerr != nil {
    testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
  }

  params := &RMRParams{&xapp.RMRParams{}}
  params.Mtype = xapp.RIC_SUB_RESP
  //params.SubId = msg.SubId
  params.SubId = -1
  params.Payload = packedMsg.Buf
  params.Meid = msg.Meid
  //params.Xid = msg.Xid
  params.Mbuf = nil

  snderr := e2termConn.RmrSend(params)
  if snderr != nil {
    testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
  }
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type test_subs_fail_params struct {
  req  *e2ap.E2APSubscriptionRequest
  fail *e2ap.E2APSubscriptionFailure
}

func (p *test_subs_fail_params) Set(req *e2ap.E2APSubscriptionRequest) {
  p.req = req

  p.fail = &e2ap.E2APSubscriptionFailure{}
  p.fail.RequestId.Id = p.req.RequestId.Id
  p.fail.RequestId.Seq = p.req.RequestId.Seq
  p.fail.FunctionId = p.req.FunctionId
  p.fail.ActionNotAdmittedList.Items = make([]e2ap.ActionNotAdmittedItem, len(p.req.ActionSetups))
  for index := int(0); index < len(p.fail.ActionNotAdmittedList.Items); index++ {
    p.fail.ActionNotAdmittedList.Items[index].ActionId = p.req.ActionSetups[index].ActionId
    p.SetCauseVal(index, 5, 1)
  }
}

func (p *test_subs_fail_params) SetCauseVal(ind int, content uint8, causeval uint8) {

  if ind < 0 {
    for index := int(0); index < len(p.fail.ActionNotAdmittedList.Items); index++ {
      p.fail.ActionNotAdmittedList.Items[index].Cause.Content = content
      p.fail.ActionNotAdmittedList.Items[index].Cause.CauseVal = causeval
    }
    return
  }
  p.fail.ActionNotAdmittedList.Items[ind].Cause.Content = content
  p.fail.ActionNotAdmittedList.Items[ind].Cause.CauseVal = causeval
}

func (e2termConn *testingE2termStub) handle_e2term_subs_fail(t *testing.T, fparams *test_subs_fail_params, msg *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_subs_fail", e2termConn.desc)
  e2SubsFail := e2t_e2asnpacker.NewPackerSubscriptionFailure()

  //---------------------------------
  // e2term activity: Send Subs Fail
  //---------------------------------
  xapp.Logger.Info("(%s) Send Subs Fail", e2termConn.desc)

  e2SubsFail.Set(fparams.fail)
  xapp.Logger.Debug("%s", e2SubsFail.String())
  packerr, packedMsg := e2SubsFail.Pack(nil)
  if packerr != nil {
    testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
  }

  params := &RMRParams{&xapp.RMRParams{}}
  params.Mtype = xapp.RIC_SUB_FAILURE
  params.SubId = msg.SubId
  params.Payload = packedMsg.Buf
  params.Meid = msg.Meid
  params.Xid = msg.Xid
  params.Mbuf = nil

  snderr := e2termConn.RmrSend(params)
  if snderr != nil {
    testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
  }
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e2termConn *testingE2termStub) handle_e2term_subs_del_req(t *testing.T) (*e2ap.E2APSubscriptionDeleteRequest, *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_subs_del_req", e2termConn.desc)
  e2SubsDelReq := e2t_e2asnpacker.NewPackerSubscriptionDeleteRequest()

  //---------------------------------
  // e2term activity: Recv Subs Del Req
  //---------------------------------
  select {
  case msg := <-e2termConn.rmrConChan:
    e2termConn.DecMsgCnt()
    if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_REQ"] {
      testError(t, "(%s) Received wrong mtype expected %s got %s, error", e2termConn.desc, "RIC_SUB_DEL_REQ", xapp.RicMessageTypeToName[msg.Mtype])
    } else {
      xapp.Logger.Info("(%s) Recv Subs Del Req", e2termConn.desc)

      packedData := &packer.PackedData{}
      packedData.Buf = msg.Payload
      unpackerr := e2SubsDelReq.UnPack(packedData)
      if unpackerr != nil {
        testError(t, "(%s) RIC_SUB_DEL_REQ unpack failed err: %s", e2termConn.desc, unpackerr.Error())
      }
      geterr, req := e2SubsDelReq.Get()
      if geterr != nil {
        testError(t, "(%s) RIC_SUB_DEL_REQ get failed err: %s", e2termConn.desc, geterr.Error())
      }
      return req, msg
    }
  case <-time.After(15 * time.Second):
    testError(t, "(%s) Not Received RIC_SUB_DEL_REQ within 15 secs", e2termConn.desc)
  }
  return nil, nil
}

func handle_e2term_recv_empty() bool {
  if len(e2termConn.rmrConChan) > 0 {
    return false
  }
  return true
}

func (e2termConn *testingE2termStub) handle_e2term_subs_del_resp(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_subs_del_resp", e2termConn.desc)
  e2SubsDelResp := e2t_e2asnpacker.NewPackerSubscriptionDeleteResponse()

  //---------------------------------
  // e2term activity: Send Subs Del Resp
  //---------------------------------
  xapp.Logger.Info("(%s) Send Subs Del Resp", e2termConn.desc)

  resp := &e2ap.E2APSubscriptionDeleteResponse{}
  resp.RequestId.Id = req.RequestId.Id
  resp.RequestId.Seq = req.RequestId.Seq
  resp.FunctionId = req.FunctionId

  e2SubsDelResp.Set(resp)
  xapp.Logger.Debug("%s", e2SubsDelResp.String())
  packerr, packedMsg := e2SubsDelResp.Pack(nil)
  if packerr != nil {
    testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
  }

  params := &RMRParams{&xapp.RMRParams{}}
  params.Mtype = xapp.RIC_SUB_DEL_RESP
  params.SubId = msg.SubId
  params.Payload = packedMsg.Buf
  params.Meid = msg.Meid
  params.Xid = msg.Xid
  params.Mbuf = nil

  snderr := e2termConn.RmrSend(params)
  if snderr != nil {
    testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
  }
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e2termConn *testingE2termStub) handle_e2term_subs_del_fail(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *RMRParams) {
  xapp.Logger.Info("(%s) handle_e2term_del_subs_fail", e2termConn.desc)
  e2SubsDelFail := e2t_e2asnpacker.NewPackerSubscriptionDeleteFailure()

  //---------------------------------
  // e2term activity: Send Subs Del Fail
  //---------------------------------
  xapp.Logger.Info("(%s) Send Subs Del Fail", e2termConn.desc)

  resp := &e2ap.E2APSubscriptionDeleteFailure{}
  resp.RequestId.Id = req.RequestId.Id
  resp.RequestId.Seq = req.RequestId.Seq
  resp.FunctionId = req.FunctionId
  resp.Cause.Content = 3  // CauseMisc
  resp.Cause.CauseVal = 4 // unspecified

  e2SubsDelFail.Set(resp)
  xapp.Logger.Debug("%s", e2SubsDelFail.String())
  packerr, packedMsg := e2SubsDelFail.Pack(nil)
  if packerr != nil {
    testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
  }

  params := &RMRParams{&xapp.RMRParams{}}
  params.Mtype = xapp.RIC_SUB_DEL_FAILURE
  params.SubId = msg.SubId
  params.Payload = packedMsg.Buf
  params.Meid = msg.Meid
  params.Xid = msg.Xid
  params.Mbuf = nil

  snderr := e2termConn.RmrSend(params)
  if snderr != nil {
    testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
  }
}
