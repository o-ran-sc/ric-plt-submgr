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
)

type Base int

func dummy() {
	fmt.Println("Testing")
}

type retransEntry struct {
	restSubsId string
	startTime  time.Time
}

type duplicateCtrl struct {
	mutex      sync.Mutex
	retransMap map[string]retransEntry
	collCount  int
}

func (c *duplicateCtrl) Init() {
	c.retransMap = make(map[string]retransEntry)
}

func (c *duplicateCtrl) HasRetransmissionOngoing(restSubsId string, payload interface{}) (error, bool, string) {

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(payload); err != nil {
		fmt.Printf("Failed to encode %v\n", payload)
		return err, false, ""
	}

	hash := md5.Sum(data.Bytes())

	md5sum := hex.EncodeToString(hash[:])

	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, present := c.retransMap[md5sum]

	if present {
		fmt.Printf("Collision detected. REST subs ID %s has ongoing transaction with MD5SUM : %s\n", entry.restSubsId, md5sum)
		c.collCount++
		return nil, true, md5sum
	}

	entry = retransEntry{restSubsId: restSubsId, startTime: time.Now()}

	fmt.Printf("Added Md5SUM %s for restSubsId %s at %s\n", md5sum, entry.restSubsId, entry.startTime)

	c.retransMap[md5sum] = entry

	return nil, false, md5sum
}

func (c *duplicateCtrl) RetransmissionComplete(md5sum string) error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, present := c.retransMap[md5sum]

	if !present {
		fmt.Printf("MD5SUM : %s NOT found from table (%v)\n", md5sum, entry)
		return fmt.Errorf("Retransmission entry not found for MD5SUM %s", md5sum)
	}

	delete(c.retransMap, md5sum)

	return nil
}
