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
/*
Timer takes four parameters:
	 1) strId 			string   			'string format timerMap key'
	 2) nbrId 			int      			'numeric format timerMap key'
	 3) timerDuration	time.Duration  		'timer duration'
	 4) tryCount        uint64              'tryCount'
	 5) timerFunction	func(string, int) 	'function to be executed when timer expires'

	Timer function is put inside in-build time.AfterFunc() Go function, where it is run inside own Go routine
	when the timer expires. Timer are two key values. Both are used always, but the other one can be left
	"empty", i.e. strId = "" or  nbrId = 0. Fourth parameter is for tryCount. Fifth parameter, the timer
	function is bare function name without 	any function parameters and parenthesis! Filling first parameter
	strId with related name can improve code readability and robustness, even the numeric Id would be enough
	from functionality point of view.

	TimerStart() function starts the timer. If TimerStart() function is called again with same key values
	while earlier started timer is still in the timerMap, i.e. it has not been stopped or the timer has not
	yet expired, the old timer is deleted and new timer is started with the given time value.

	StopTimer() function stops the timer. There is no need to call StopTimer() function after the timer has
	expired. Timer is removed automatically from the timeMap. Calling StopTimer() function with key values not
	existing in the timerMap, has no effect.

	NOTE: Each timer is run in separate Go routine. Therefore, the function that is executed when timer expires
	MUST be designed to be able run concurrently! Also, function run order of simultaneously expired timers cannot
	guaranteed anyway!

	If you need to transport more information to the timer function, consider to use another map to store the
	information with same key value, as the started timer.

	Init timerMap example:
		timerMap := new(TimerMap)
		timerMap.Init()

	StartTimer() and StartTimer() function usage examples.
	1)
		subReqTime := 2 * time.Second
		subId := 123
		var tryCount uint64 = 1
		timerMap.StartTimer("RIC_SUB_REQ", int(subId), subReqTime, FirstTry, handleSubscriptionRequestTimer)
		timerMap.StopTimer("RIC_SUB_REQ", int(subId))


	StartTimer() retry example.
	2)
		subReqTime := 2 * time.Second
		subId := 123
		var tryCount uint64 = 1
		timerMap.StartTimer("RIC_SUB_REQ", int(subId), subReqTime, FirstTry, handleSubscriptionRequestTimer)
		timerMap.StopTimer("RIC_SUB_REQ", int(subId))

	3)
		subReqTime := 2 * time.Second
		strId := "1UHSUwNqxiVgUWXvC4zFaatpZFF"
		var tryCount uint64 = 1
		timerMap.StartTimer(strId, 0, subReqTime, FirstTry, handleSubscriptionRequestTimer)
		timerMap.StopTimer(strId, 0)

	4)
		subReqTime := 2 * time.Second
		strId := "1UHSUwNqxiVgUWXvC4zFaatpZFF"
		var tryCount uint64 = 1
		timerMap.StartTimer(RIC_SUB_REQ_" + strId, 0, subReqTime, FirstTry, handleSubscriptionRequestTimer)
		timerMap.timerMap.StopTimer("RIC_SUB_REQ_" + strId, 0)

	Timer function example. This is run if any of the above started timer expires.
		func handleSubscriptionRequestTimer1(strId string, nbrId int, tryCount uint64) {
			fmt.Printf("Subscription Request timer expired. Name: %v, SubId: %v, tryCount: %v\n",strId, nbrId, tryCount)
			...

			// Retry
			....

			tryCount++
		    timerMap.StartTimer("RIC_SUB_REQ", int(subId), subReqTime, tryCount, handleSubscriptionRequestTimer)
			...

		}
*/

package control

import (
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"sync"
	"time"
)

const FirstTry = 1

type TimerKey struct {
	strId string
	nbrId int
}

type TimerInfo struct {
	timerAddress         *time.Timer
	timerFunctionAddress func()
}

type TimerMap struct {
	timer map[TimerKey]TimerInfo
	mutex sync.Mutex
}

// This method should run as a constructor
func (t *TimerMap) Init() {
	t.timer = make(map[TimerKey]TimerInfo)
}

func (t *TimerMap) StartTimer(strId string, nbrId int, expireAfterTime time.Duration, tryCount uint64, timerFunction func(srtId string, nbrId int, tryCount uint64)) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if timerFunction == nil {
		xapp.Logger.Error("StartTimer() timerFunc == nil\n")
		return false
	}
	timerKey := TimerKey{strId, nbrId}
	// Stop timer if there is already timer running with the same id
	if val, ok := t.timer[timerKey]; ok {
		xapp.Logger.Debug("StartTimer() old timer found")
		if val.timerAddress != nil {
			xapp.Logger.Debug("StartTimer() deleting old timer")
			val.timerAddress.Stop()
		}
		delete(t.timer, timerKey)
	}

	// Store in timerMap in-build Go "timer", timer function executor and the function to be executed when the timer expires
	t.timer[timerKey] = TimerInfo{timerAddress: time.AfterFunc(expireAfterTime, func() { t.timerFunctionExecutor(strId, nbrId) }),
		timerFunctionAddress: func() { timerFunction(strId, nbrId, tryCount) }}
	return true
}

func (t *TimerMap) StopTimer(strId string, nbrId int) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	timerKey := TimerKey{strId, nbrId}
	if val, ok := t.timer[timerKey]; ok {
		if val.timerAddress != nil {
			val.timerAddress.Stop()
			delete(t.timer, timerKey)
			return true
		} else {
			xapp.Logger.Error("StopTimer() timerAddress == nil")
			return false
		}
	} else {
		xapp.Logger.Debug("StopTimer() Timer not found. May be expired or stopped already. timerKey.strId: %v, timerKey.strId: %v\n", timerKey.strId, timerKey.nbrId)
		return false
	}
}

func (t *TimerMap) timerFunctionExecutor(strId string, nbrId int) {
	t.mutex.Lock()
	timerKey := TimerKey{strId, nbrId}
	if val, ok := t.timer[timerKey]; ok {
		if val.timerFunctionAddress != nil {
			// Take local copy of timer function address
			f := val.timerFunctionAddress
			// Delete timer instance from map
			delete(t.timer, timerKey)
			t.mutex.Unlock()
			// Execute the timer function
			f()
			return
		} else {
			xapp.Logger.Error("timerExecutorFunc() timerFunctionAddress == nil")
			t.mutex.Unlock()
			return
		}
	} else {
		xapp.Logger.Error("timerExecutorFunc() Timer is not anymore in map. timerKey.strId: %v, timerKey.strId: %v\n", timerKey.strId, timerKey.nbrId)
		t.mutex.Unlock()
		return
	}
}
