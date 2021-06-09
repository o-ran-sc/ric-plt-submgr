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
	"fmt"
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	Name    *string
	Data    []byte
	SomeVal *int64
}

func TestDefaultUseCase(t *testing.T) {

	fmt.Println("#####################  TestRetransmissionChecker  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}

	retransCtrl.Init()

	_, duplicate, md5sum := retransCtrl.IsDuplicateToOngoingTransaction(restSubdId, data)

	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, false, duplicate)

	retransCtrl.TransactionComplete(md5sum)

	assert.Equal(t, 0, len(retransCtrl.retransMap))
}

func TestDuplicate(t *testing.T) {

	fmt.Println("#####################  TestDuplicate  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}

	var name2 string = "yolo"
	var someVal2 int64 = 98765

	data2 := new(testData)
	data2.Name = &name2
	data2.SomeVal = &someVal2
	datax := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	data2.Data = datax

	retransCtrl.Init()

	_, duplicate, md5sum := retransCtrl.IsDuplicateToOngoingTransaction(restSubdId, data)
	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, false, duplicate)

	_, duplicate, md5sum = retransCtrl.IsDuplicateToOngoingTransaction(restSubdId, data2)
	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, true, duplicate)

	retransCtrl.TransactionComplete(md5sum)

	assert.Equal(t, 0, len(retransCtrl.retransMap))
}

func TestEncodingError(t *testing.T) {

	fmt.Println("#####################  TestEncodingError  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var data interface{}

	retransCtrl.Init()

	err, duplicate, _ := retransCtrl.IsDuplicateToOngoingTransaction(restSubdId, data)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, 0, len(retransCtrl.retransMap))
	assert.Equal(t, false, duplicate)
}

func TestRemovalError(t *testing.T) {

	fmt.Println("#####################  TestRemovalError  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var data testData

	retransCtrl.Init()

	err, duplicate, md5sum := retransCtrl.IsDuplicateToOngoingTransaction(restSubdId, data)
	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, false, duplicate)

	err = retransCtrl.TransactionComplete(md5sum)
	assert.Empty(t, err)

	err = retransCtrl.TransactionComplete(md5sum)
	assert.NotEmpty(t, err)
}

func TestXappRestReqDuplicate(t *testing.T) {

	fmt.Println("#####################  TestXappRestReqDuplicate  #####################")

	var retransCtrl duplicateCtrl

	msg1 := new(models.SubscriptionParams)
	msg2 := new(models.SubscriptionParams)

	retransCtrl.Init()

	_, duplicate, md5sum := retransCtrl.IsDuplicateToOngoingTransaction("foobar", msg1)
	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, false, duplicate)

	_, duplicate, md5sum = retransCtrl.IsDuplicateToOngoingTransaction("foobar", msg2)
	assert.Equal(t, 1, len(retransCtrl.retransMap))
	assert.Equal(t, true, duplicate)

	retransCtrl.TransactionComplete(md5sum)

	assert.Equal(t, 0, len(retransCtrl.retransMap))
}
