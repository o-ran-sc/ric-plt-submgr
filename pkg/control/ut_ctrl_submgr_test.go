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
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingSubmgrControl struct {
	teststub.RmrControl
	c *Control
}

type Counter struct {
	Name  string
	Value uint64
}

type CountersToBeAdded []Counter

var countersBeforeMap map[string]Counter
var toBeAddedCountersMap map[string]Counter

func createSubmgrControl(srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc) *testingSubmgrControl {
	mainCtrl = &testingSubmgrControl{}
	mainCtrl.RmrControl.Init("SUBMGRCTL", srcId, rtgSvc)
	mainCtrl.c = NewControl()
	xapp.Logger.Debug("Replacing real db with test db")
	mainCtrl.c.db = CreateMock() // This overrides real database for testing
	xapp.SetReadyCB(mainCtrl.ReadyCB, nil)
	go xapp.RunWithParams(mainCtrl.c, false)
	mainCtrl.WaitCB()
	mainCtrl.c.ReadyCB(nil)
	return mainCtrl
}

func (mc *testingSubmgrControl) SimulateRestart(t *testing.T) {
	mc.TestLog(t, "Simulating submgr restart")
	mainCtrl.c.registry.subIds = nil
	// Initialize subIds slice and subscription map
	mainCtrl.c.registry.Initialize()
	// Read subIds and subscriptions from database
	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		mc.TestError(t, "%v", err)
	} else {
		mainCtrl.c.registry.register = nil
		mainCtrl.c.registry.subIds = subIds
		mainCtrl.c.registry.register = register

		mc.TestLog(t, "register:")
		for subId, subs := range register {
			mc.TestLog(t, "  subId=%v", subId)
			mc.TestLog(t, "  subs.SubRespRcvd=%v", subs.SubRespRcvd)
			mc.TestLog(t, "  subs=%v\n", subs)
		}

		mc.TestLog(t, "mainCtrl.c.registry.register:")
		for subId, subs := range mainCtrl.c.registry.register {
			mc.TestLog(t, "  subId=%v", subId)
			mc.TestLog(t, "  subs.SubRespRcvd=%v", subs.SubRespRcvd)
			mc.TestLog(t, "  subs=%v\n", subs)
		}
	}
	go mainCtrl.c.HandleUncompletedSubscriptions(mainCtrl.c.registry.register)
}

func (mc *testingSubmgrControl) SetResetTestFlag(t *testing.T, status bool) {
	mc.TestLog(t, "ResetTestFlag set to %v=", status)
	mainCtrl.c.ResetTestFlag = status
}

func (mc *testingSubmgrControl) removeExistingSubscriptions(t *testing.T) {

	mc.TestLog(t, "Removing existing subscriptions")
	mainCtrl.c.RemoveAllSubscriptionsFromSdl()
	mainCtrl.c.registry.subIds = nil
	// Initialize subIds slice and subscription map
	mainCtrl.c.registry.Initialize()
}

func PringSubscriptionQueryResult(resp models.SubscriptionList) {
	for _, item := range resp {
		fmt.Printf("item.SubscriptionID=%v\n", item.SubscriptionID)
		fmt.Printf("item.Meid=%v\n", item.Meid)
		fmt.Printf("item.ClientEndpoint=%v\n", item.ClientEndpoint)
	}
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

func (mc *testingSubmgrControl) wait_multi_subs_clean(t *testing.T, e2SubsIds []uint32, secs int) bool {
	var subs *Subscription
	var purgedSubscriptions int
	i := 1
	k := 0
	for ; i <= secs*2; i++ {
		purgedSubscriptions = 0
		for k = 0; k <= len(e2SubsIds); i++ {
			subs = mc.c.registry.GetSubscription(e2SubsIds[k])
			if subs == nil {
				mc.TestLog(t, "(submgr) subscriber purged for esSubsId %v", e2SubsIds[k])
				purgedSubscriptions += 1
				if purgedSubscriptions == len(e2SubsIds) {
					return true
				} else {
					continue
				}
			} else {
				mc.TestLog(t, "(submgr) subscriber %s no clean within %d secs: subs(N/A) - purged subscriptions %v", subs.String(), secs, purgedSubscriptions)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	mc.TestError(t, "(submgr) no clean within %d secs: subs(N/A) - %v/%v subscriptions found still", secs, purgedSubscriptions, len(e2SubsIds))

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

func (mc *testingSubmgrControl) GetMetrics(t *testing.T) (string, error) {
	req, err := http.NewRequest("GET", "http://localhost:8080/ric/v1/metrics", nil)
	if err != nil {
		return "", fmt.Errorf("Error reading request. %v", err)
	}
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error reading response. %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading body. %v", err)
	}
	return string(respBody[:]), nil
}

func (mc *testingSubmgrControl) CounterValuesToBeVeriefied(t *testing.T, countersToBeAdded CountersToBeAdded) {

	if len(toBeAddedCountersMap) == 0 {
		toBeAddedCountersMap = make(map[string]Counter)
	}
	for _, counter := range countersToBeAdded {
		toBeAddedCountersMap[counter.Name] = counter
	}
	mc.GetCounterValuesBefore(t)
}

func (mc *testingSubmgrControl) GetCounterValuesBefore(t *testing.T) {
	countersBeforeMap = make(map[string]Counter)
	countersBeforeMap = mc.GetCurrentCounterValues(t, toBeAddedCountersMap)
}

func (mc *testingSubmgrControl) VerifyCounterValues(t *testing.T) {
	currentCountersMap := mc.GetCurrentCounterValues(t, toBeAddedCountersMap)
	for _, toBeAddedCounter := range toBeAddedCountersMap {
		if currentCounter, ok := currentCountersMap[toBeAddedCounter.Name]; ok == true {
			if beforeCounter, ok := countersBeforeMap[toBeAddedCounter.Name]; ok == true {
				if currentCounter.Value != beforeCounter.Value+toBeAddedCounter.Value {
					mc.TestError(t, "Error in expected counter value: counterName %v, current value %v, expected value %v",
						currentCounter.Name, currentCounter.Value, beforeCounter.Value+toBeAddedCounter.Value)

					//fmt.Printf("beforeCounter.Value=%v, toBeAddedCounter.Value=%v, \n",beforeCounter.Value, toBeAddedCounter.Value)
				}
			} else {
				mc.TestError(t, "Counter %v not in countersBeforeMap", toBeAddedCounter.Name)
			}
		} else {
			mc.TestError(t, "Counter %v not in currentCountersMap", toBeAddedCounter.Name)
		}
	}

	// Make map empty
	//fmt.Printf("toBeAddedCountersMap=%v\n",toBeAddedCountersMap)
	toBeAddedCountersMap = make(map[string]Counter)
}

func (mc *testingSubmgrControl) GetCurrentCounterValues(t *testing.T, chekedCountersMap map[string]Counter) map[string]Counter {
	countersString, err := mc.GetMetrics(t)
	if err != nil {
		mc.TestError(t, "Error GetMetrics() failed %v", err)
		return nil
	}

	retCounterMap := make(map[string]Counter)
	stringsTable := strings.Split(countersString, "\n")
	for _, counter := range chekedCountersMap {
		for _, counterString := range stringsTable {
			if !strings.Contains(counterString, "#") && strings.Contains(counterString, counter.Name) {
				counterString := strings.Split(counterString, " ")
				if strings.Contains(counterString[0], counter.Name) {
					val, err := strconv.ParseUint(counterString[1], 10, 64)
					if err != nil {
						mc.TestError(t, "Error: strconv.ParseUint failed %v", err)
					}
					counter.Value = val
					//fmt.Printf("counter=%v\n", counter)
					retCounterMap[counter.Name] = counter
				}
			}
		}
	}

	if len(retCounterMap) != len(chekedCountersMap) {
		mc.TestError(t, "Error: len(retCounterMap) != len(chekedCountersMap)")

	}
	return retCounterMap
}

func (mc *testingSubmgrControl) sendGetRequest(t *testing.T, addr string, path string) {

	mc.TestLog(t, "GET http://"+addr+"%v", path)
	req, err := http.NewRequest("GET", "http://"+addr+path, nil)
	if err != nil {
		mc.TestError(t, "Error reading request. %v", err)
		return
	}
	req.Header.Set("Cache-Control", "no-cache")
	client := &http.Client{Timeout: time.Second * 2}
	resp, err := client.Do(req)
	if err != nil {
		mc.TestError(t, "Error reading response. %v", err)
		return
	}
	defer resp.Body.Close()

	mc.TestLog(t, "Response status: %v", resp.Status)
	mc.TestLog(t, "Response Headers: %v", resp.Header)
	if !strings.Contains(resp.Status, "200 OK") {
		mc.TestError(t, "Wrong response status")
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mc.TestError(t, "Error reading body. %v", err)
		return
	}
	mc.TestLog(t, "%s", respBody)
	return
}

func (mc *testingSubmgrControl) sendPostRequest(t *testing.T, addr string, path string) {

	mc.TestLog(t, "POST http://"+addr+"%v", path)
	req, err := http.NewRequest("POST", "http://"+addr+path, nil)
	if err != nil {
		mc.TestError(t, "Error reading request. %v", err)
		return
	}
	client := &http.Client{Timeout: time.Second * 2}
	resp, err := client.Do(req)
	if err != nil {
		mc.TestError(t, "Error reading response. %v", err)
		return
	}
	defer resp.Body.Close()

	mc.TestLog(t, "Response status: %v", resp.Status)
	mc.TestLog(t, "Response Headers: %v", resp.Header)
	if !strings.Contains(resp.Status, "200 OK") {
		mc.TestError(t, "Wrong response status")
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mc.TestError(t, "Error reading body. %v", err)
		return
	}
	mc.TestLog(t, "%s", respBody)
	return
}
