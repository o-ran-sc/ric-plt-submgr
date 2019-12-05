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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"sync"
	"time"
)

var timerMutex = &sync.Mutex{}

type TimerInfo struct {
	timerAddress *time.Timer	
	timerFunctionAddress func()
}

type TimerMap struct {
	timer map[uint16] TimerInfo
}

// This method should run as a constructor
func (t *TimerMap) Init() {
	t.timer = make(map[uint16] TimerInfo)
}

func (t *TimerMap) StartTimer(subId uint16, expireAfterTime time.Duration, timerFunction func(subId uint16)) bool {
	timerMutex.Lock()
	defer timerMutex.Unlock()
	if (timerFunction == nil) {
		xapp.Logger.Error("StartTimer() timerFunc == nil")
		return false
	}

	// Stop timer if there is already timer running with the same id
	if val, ok := t.timer[subId]; ok {
		xapp.Logger.Error("StartTimer() old timer found")
		if val.timerAddress != nil {
			xapp.Logger.Error("StartTimer() deleting old timer")
			val.timerAddress.Stop()
		}
		delete(t.timer, subId)
	}

	// Store timer + timer function excecutor function and the function to be excecuted when timer expires, in map
	t.timer[subId] = TimerInfo{timerAddress: time.AfterFunc(expireAfterTime, func(){t.timerFunctionExcecutor(subId)}),
							   timerFunctionAddress: func(){timerFunction(subId)}}
	return true
}

func (t *TimerMap) StopTimer(subId uint16) bool {
	timerMutex.Lock()
	defer timerMutex.Unlock()
	if val, ok := t.timer[subId]; ok {
		if val.timerAddress != nil {
			val.timerAddress.Stop()
			delete(t.timer, subId)
			return true
		} else {
			xapp.Logger.Error("StopTimer() timerAddress == nil")
			return false
		}
	} else {
		xapp.Logger.Info("StopTimer() Timer not found. May be expired or stopped already. subId: %v",subId)
		return false
	}
}

func (t *TimerMap) timerFunctionExcecutor(subId uint16) {
	timerMutex.Lock()
	if val, ok := t.timer[subId]; ok {
		if val.timerFunctionAddress != nil {
			// Take local copy of timer function address
			f := val.timerFunctionAddress
			// Delete timer instance from map
			delete(t.timer, subId)
			timerMutex.Unlock()
			// Excecute the timer function
			f()
			return
		} else {
			xapp.Logger.Error("timerExcecutorFunc() timerFunctionAddress == nil")
			timerMutex.Unlock()
			return
		}
	} else {
		xapp.Logger.Error("timerExcecutorFunc() Timer not anymore in map. subId: %v",subId)
		timerMutex.Unlock()
		return
	}
}
