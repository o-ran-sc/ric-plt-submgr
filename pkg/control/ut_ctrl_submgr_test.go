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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingSubmgrControl struct {
	teststub.RmrControl
	c *Control
}

func createSubmgrControl(srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc) *testingSubmgrControl {
	mainCtrl = &testingSubmgrControl{}
	mainCtrl.RmrControl.Init("SUBMGRCTL", srcId, rtgSvc)
	mainCtrl.c = NewControl()
	xapp.SetReadyCB(mainCtrl.ReadyCB, nil)
	go xapp.RunWithParams(mainCtrl.c, false)
	mainCtrl.WaitCB()
	mainCtrl.c.ReadyCB(nil)
	return mainCtrl
}

func (mc *testingSubmgrControl) wait_registry_empty(t *testing.T, secs int) bool {
	cnt := int(0)
	i := 1
	for ; i <= secs*2; i++ {
		cnt = len(mc.c.registry.register)
		if cnt == 0 {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	mc.TestError(t, "(submgr) no registry empty within %d secs: %d", secs, cnt)
	return false
}

func (mc *testingSubmgrControl) get_registry_next_subid(t *testing.T) uint32 {
	mc.c.registry.mutex.Lock()
	defer mc.c.registry.mutex.Unlock()
	return mc.c.registry.subIds[0]
}

func (mc *testingSubmgrControl) wait_registry_next_subid_change(t *testing.T, origSubId uint32, secs int) (uint32, bool) {
	i := 1
	for ; i <= secs*2; i++ {
		mc.c.registry.mutex.Lock()
		currSubId := mc.c.registry.subIds[0]
		mc.c.registry.mutex.Unlock()
		if currSubId != origSubId {
			return currSubId, true
		}
		time.Sleep(500 * time.Millisecond)
	}
	mc.TestError(t, "(submgr) no subId change within %d secs", secs)
	return 0, false
}

func (mc *testingSubmgrControl) wait_subs_clean(t *testing.T, e2SubsId uint32, secs int) bool {
	var subs *Subscription
	i := 1
	for ; i <= secs*2; i++ {
		subs = mc.c.registry.GetSubscription(e2SubsId)
		if subs == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	if subs != nil {
		mc.TestError(t, "(submgr) no clean within %d secs: %s", secs, subs.String())
	} else {
		mc.TestError(t, "(submgr) no clean within %d secs: subs(N/A)", secs)
	}
	return false
}

func (mc *testingSubmgrControl) wait_subs_trans_clean(t *testing.T, e2SubsId uint32, secs int) bool {
	var trans TransactionIf
	i := 1
	for ; i <= secs*2; i++ {
		subs := mc.c.registry.GetSubscription(e2SubsId)
		if subs == nil {
			return true
		}
		trans = subs.GetTransaction()
		if trans == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	if trans != nil {
		mc.TestError(t, "(submgr) no clean within %d secs: %s", secs, trans.String())
	} else {
		mc.TestError(t, "(submgr) no clean within %d secs: trans(N/A)", secs)
	}
	return false
}

func (mc *testingSubmgrControl) get_subs_entrypoint_cnt(t *testing.T, origSubId uint32) int {
	subs := mc.c.registry.GetSubscription(origSubId)
	if subs == nil {
		mc.TestError(t, "(submgr) no subs %d exists during entrypoint cnt get", origSubId)
		return -1
	}
	return subs.EpList.Size()
}

func (mc *testingSubmgrControl) wait_subs_entrypoint_cnt_change(t *testing.T, origSubId uint32, orig int, secs int) (int, bool) {

	subs := mc.c.registry.GetSubscription(origSubId)
	if subs == nil {
		mc.TestError(t, "(submgr) no subs %d exists during entrypoint cnt wait", origSubId)
		return -1, true
	}

	i := 1
	for ; i <= secs*2; i++ {
		curr := subs.EpList.Size()
		if curr != orig {
			return curr, true
		}
		time.Sleep(500 * time.Millisecond)
	}
	mc.TestError(t, "(submgr) no subs %d entrypoint cnt change within %d secs", origSubId, secs)
	return 0, false
}

//
// Counter check for received message. Note might not be yet handled
//
func (mc *testingSubmgrControl) get_msgcounter(t *testing.T) uint64 {
	return mc.c.CntRecvMsg
}

func (mc *testingSubmgrControl) wait_msgcounter_change(t *testing.T, orig uint64, secs int) (uint64, bool) {
	i := 1
	for ; i <= secs*2; i++ {
		curr := mc.c.CntRecvMsg
		if curr != orig {
			return curr, true
		}
		time.Sleep(500 * time.Millisecond)
	}
	mc.TestError(t, "(submgr) no msg counter change within %d secs", secs)
	return 0, false
}
