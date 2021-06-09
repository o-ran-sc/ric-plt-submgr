/*
==================================================================================
  Copyright (c) 2021 Nokia

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
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type retransEntry struct {
	restSubsId string
	startTime  time.Time
}

type duplicateCtrl struct {
	mutex      sync.Mutex
	retransMap map[string]retransEntry
	collCount  int
}

func (d *duplicateCtrl) Init() {
	d.retransMap = make(map[string]retransEntry)
}

func (d *duplicateCtrl) IsDuplicateToOngoingTransaction(restSubsId string, payload interface{}) (error, bool, string) {

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(payload); err != nil {
		xapp.Logger.Error("Failed to encode %v\n", payload)
		return err, false, ""
	}

	hash := md5.Sum(data.Bytes())

	md5sum := hex.EncodeToString(hash[:])

	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, present := d.retransMap[md5sum]

	if present {
		xapp.Logger.Info("Collision detected. REST subs ID %s has ongoing transaction with MD5SUM : %s started at %s\n", entry.restSubsId, md5sum, entry.startTime.Format(time.ANSIC))
		d.collCount++
		return nil, true, md5sum
	}

	entry = retransEntry{restSubsId: restSubsId, startTime: time.Now()}

	xapp.Logger.Debug("Added Md5SUM %s for restSubsId %s at %s\n", md5sum, entry.restSubsId, entry.startTime)

	d.retransMap[md5sum] = entry

	return nil, false, md5sum
}

func (d *duplicateCtrl) TransactionComplete(md5sum string) error {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, present := d.retransMap[md5sum]

	if !present {
		xapp.Logger.Error("MD5SUM : %s NOT found from table (%v)\n", md5sum, entry)
		return fmt.Errorf("Retransmission entry not found for MD5SUM %s", md5sum)
	}

	xapp.Logger.Debug("Releasing transaction duplicate blocker for %s, MD5SUM : %s\n", entry.restSubsId, md5sum)

	delete(d.retransMap, md5sum)

	return nil
}
