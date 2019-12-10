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

package conv

import (
	"os"
	"testing"
)

// Test cases
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestBcdEven(t *testing.T) {

	bcd := NewBcd("0123456789??????")
	bcdbuf := bcd.Encode("123456")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdEven: bcd Encode failed")
	}

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("123456") {
		t.Errorf("TestBcdEven: bcd Decode failed: got %s expect %s", bcdstr, string("123456"))
	}

}

func TestBcdUnEven1(t *testing.T) {

	bcd := NewBcd("0123456789??????")
	bcdbuf := bcd.Encode("12345")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdUnEven1: bcd Encode failed")
	}

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345?") {
		t.Errorf("TestBcdUnEven1: bcd Decode failed: got %s expect %s", bcdstr, string("12345?"))
	}
}

func TestBcdUnEven2(t *testing.T) {

	bcd := NewBcd("0123456789?????f")
	bcdbuf := bcd.Encode("12345f")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdUnEven2: bcd Encode failed")
	}

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345f") {
		t.Errorf("TestBcdUnEven2: bcd Decode failed: got %s expect %s", bcdstr, string("12345f"))
	}
}
