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

type testData struct {
	Name    *string
	Data    []byte
	SomeVal *int64
}

/*
func TestRetransmissionChecker(t *testing.T) {

	fmt.Println("#####################  TestRetransmissionChecker  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}

	retransCtrl.Init(100*time.Millisecond, 300*time.Millisecond)

	err, _, md5sum := retransCtrl.HasRetransmissionOngoing(restSubdId, data)

	assert.Equal(t, err, nil)

	time.Sleep(time.Duration(time.Millisecond * 200))

	err = retransCtrl.RetransmissionComplete(md5sum)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	time.Sleep(time.Duration(time.Millisecond * 500))

	retransCtrl.Shutdown()
}

func TestBlockRetranmission(t *testing.T) {

	fmt.Println("#####################  TestBlockRetranmission  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"

	var name string = "yolo"
	var name2 string = "yolo"
	var someVal int64 = 98765
	var someVal2 int64 = 98765

	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}
	data2 := testData{Name: &name2, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal2}

	retransCtrl.Init(100*time.Millisecond, 300*time.Millisecond)

	err, ongoing, md5sum := retransCtrl.HasRetransmissionOngoing(restSubdId, data)

	assert.Equal(t, err, nil)
	assert.Equal(t, ongoing, false)

	time.Sleep(time.Duration(time.Millisecond * 300))

	err, ongoing, md5sum = retransCtrl.HasRetransmissionOngoing(restSubdId, data2)

	assert.Equal(t, err, nil)
	assert.Equal(t, ongoing, true)

	time.Sleep(time.Duration(time.Millisecond * 200))

	err = retransCtrl.RetransmissionComplete(md5sum)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	time.Sleep(time.Duration(time.Millisecond * 500))

	retransCtrl.Shutdown()
	time.Sleep(time.Duration(time.Millisecond * 500))
}

func TestBlockRetranmissionDuringGuarantinePeriod(t *testing.T) {

	fmt.Println("#####################  TestBlockRetranmissionDuringGuarantinePeriod  #####################")

	var retransCtrl duplicateCtrl
	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}

	retransCtrl.Init(100*time.Millisecond, 300*time.Millisecond)

	err, ongoing, md5sum := retransCtrl.HasRetransmissionOngoing(restSubdId, data)

	assert.Equal(t, err, nil)
	assert.Equal(t, ongoing, false)

	time.Sleep(time.Duration(time.Millisecond * 300))

	err = retransCtrl.RetransmissionComplete(md5sum)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	time.Sleep(time.Duration(time.Millisecond * 100))

	err, ongoing, md5sum = retransCtrl.HasRetransmissionOngoing(restSubdId, data)

	assert.Equal(t, err, nil)
	assert.Equal(t, ongoing, true)

	time.Sleep(time.Duration(time.Millisecond * 300))

	retransCtrl.Shutdown()
}

func TestBlockRetranmissionThreeConcurrentReqs(t *testing.T) {

	fmt.Println("#####################  TestBlockRetranmissionThreeConcurrentReqs  #####################")

	var retransCtrl duplicateCtrl
	retransCtrl.Init(100*time.Millisecond, 300*time.Millisecond)

	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}
	var md5sums [3]string
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {

		wg.Add(1)
		go func(i int, data interface{}) {

			d := data.(testData)
			d.Data[len(d.Data)-1] = (byte)(i)

			err, ongoing, md5sum := retransCtrl.HasRetransmissionOngoing(restSubdId+strconv.Itoa(i), data)

			md5sums[i] = md5sum

			assert.Equal(t, err, nil)
			assert.Equal(t, ongoing, false)

			time.Sleep(time.Duration(time.Millisecond * 300))
			defer wg.Done()
		}(i, data)
	}

	wg.Wait()

	for i := 0; i < 3; i++ {

		wg.Add(1)
		go func(i int) {

			err := retransCtrl.RetransmissionComplete(md5sums[i])

			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}

			defer wg.Done()
		}(i)
	}

	time.Sleep(time.Duration(time.Millisecond * 1000))
	wg.Wait()
}

func TestBlockRetranmissionWithCOncurrentReqs(t *testing.T) {

	fmt.Println("#####################  TestBlockRetranmissionWithCOncurrentReqs  #####################")
	var retransCtrl duplicateCtrl
	retransCtrl.Init(100*time.Millisecond, 300*time.Millisecond)

	restSubdId := "898dfkjashntgkjasgho4"
	var name string = "yolo"
	var someVal int64 = 98765
	data := testData{Name: &name, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, SomeVal: &someVal}
	var md5sums [3]string
	var wg sync.WaitGroup

	for k := 0; k < 3; k++ {

		wg.Add(1)
		go func(i int, data interface{}) {

			d := data.(testData)
			d.Data[len(d.Data)-1] = (byte)(i % 2)

			_, _, md5sum := retransCtrl.HasRetransmissionOngoing(restSubdId+strconv.Itoa(i%2), data)

			md5sums[i] = md5sum

			time.Sleep(time.Duration(time.Millisecond * 300))
			defer wg.Done()
		}(k, data)
	}

	wg.Wait()

	assert.Equal(t, retransCtrl.collCount, 1)

	for i := 0; i < 3; i++ {

		wg.Add(1)
		go func(i int) {

			err := retransCtrl.RetransmissionComplete(md5sums[i])

			if i < 3 && err != nil {
				t.Error("Retransmission complete failure")
			}

			defer wg.Done()
		}(i)
	}

	wg.Wait()
	time.Sleep(time.Duration(time.Millisecond * 1000))
}
*/
