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

//-----------------------------------------------------------------------------
//
// MCC 3 digits MNC 2 digits
// BCD Coded format: 0xC2C1 0xfC3 0xN2N1
// String format   : C1C2C3N1N2
//
// MCC 3 digits MNC 3 digits
// BCD Coded format: 0xC2C1 0xN3C3 0xN2N1
// String format   : C1C2C3N1N2N3
//
//-----------------------------------------------------------------------------

type PlmnIdentity struct {
	Val [3]uint8
}

func (plmnid *PlmnIdentity) String() string {
	bcd := NewBcd("0123456789?????f")

	str := bcd.Decode(plmnid.Val[:])

	if str[3] == 'f' {
		return string(str[0:3]) + string(str[4:])
	}
	return string(str[0:3]) + string(str[4:]) + string(str[3])
}

func (plmnid *PlmnIdentity) MccString() string {
	fullstr := plmnid.String()
	return string(fullstr[0:3])
}

func (plmnid *PlmnIdentity) MncString() string {
	fullstr := plmnid.String()
	return string(fullstr[3:])
}

func (plmnid *PlmnIdentity) StringPut(str string) bool {

	var tmpStr string
	switch {

	case len(str) == 5:
		//C1 C2 C3 N1 N2 -->
		//C2C1 0fC3 N2N1
		tmpStr = string(str[0:3]) + string("f") + string(str[3:])
	case len(str) == 6:
		//C1 C2 C3 N1 N2 N3 -->
		//C2C1 N3C3 N2N1
		tmpStr = string(str[0:3]) + string(str[5]) + string(str[3:5])
	default:
		return false
	}

	bcd := NewBcd("0123456789?????f")
	buf := bcd.Encode(tmpStr)

	if buf == nil {
		return false
	}

	return plmnid.BcdPut(buf)
}

func (plmnid *PlmnIdentity) BcdPut(val []uint8) bool {

	if len(val) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		plmnid.Val[i] = val[i]
	}
	return true
}
