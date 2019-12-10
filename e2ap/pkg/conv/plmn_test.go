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
	"testing"
)

func TestPlmnId1(t *testing.T) {

	var ident PlmnIdentity
	ident.StringPut("23350")

	if ident.Val[0] != 0x32 {
		t.Errorf("TestPlmnId1: ident.val[0] expected 0x32 got 0x%x", ident.Val[0])
	}

	if ident.Val[1] != 0xf3 {
		t.Errorf("TestPlmnId1: ident.val[1] expected 0xf3 got 0x%x", ident.Val[1])
	}

	if ident.Val[2] != 0x05 {
		t.Errorf("TestPlmnId1: ident.val[2] expected 0x05 got 0x%x", ident.Val[2])
	}

	fullstr := ident.String()
	if fullstr != "23350" {
		t.Errorf("TestPlmnId2: fullstr expected 23350 got %s", fullstr)
	}

	mccstr := ident.MccString()
	if mccstr != "233" {
		t.Errorf("TestPlmnId1: mcc expected 233 got %s", mccstr)
	}
	mncstr := ident.MncString()
	if mncstr != "50" {
		t.Errorf("TestPlmnId1: mnc expected 50 got %s", mncstr)
	}
}

func TestPlmnId2(t *testing.T) {

	var ident PlmnIdentity
	ident.StringPut("233550")

	if ident.Val[0] != 0x32 {
		t.Errorf("TestPlmnId1: ident.val[0] expected 0x32 got 0x%x", ident.Val[0])
	}

	if ident.Val[1] != 0x03 {
		t.Errorf("TestPlmnId1: ident.val[1] expected 0x03 got 0x%x", ident.Val[1])
	}

	if ident.Val[2] != 0x55 {
		t.Errorf("TestPlmnId1: ident.val[2] expected 0x55 got 0x%x", ident.Val[2])
	}

	fullstr := ident.String()
	if fullstr != "233550" {
		t.Errorf("TestPlmnId2: fullstr expected 233550 got %s", fullstr)
	}

	mccstr := ident.MccString()
	if mccstr != "233" {
		t.Errorf("TestPlmnId2: mcc expected 233 got %s", mccstr)
	}
	mncstr := ident.MncString()
	if mncstr != "550" {
		t.Errorf("TestPlmnId2: mnc expected 550 got %s", mncstr)
	}
}
