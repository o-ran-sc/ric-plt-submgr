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

type Bcd struct {
	ConvTbl string
}

func NewBcd(convTbl string) *Bcd {
	b := &Bcd{}
	b.ConvTbl = convTbl
	return b
}

func (bcd *Bcd) index(c byte) int {
	for cpos, cchar := range bcd.ConvTbl {
		if cchar == rune(c) {
			return cpos
		}
	}
	return -1
}

func (bcd *Bcd) byte(i int) byte {
	if i < 0 && i > 15 {
		return '?'
	}
	return bcd.ConvTbl[i]
}

func (bcd *Bcd) Encode(str string) []byte {
	buf := make([]byte, len(str)/2+len(str)%2)
	for i := 0; i < len(str); i++ {
		var schar int = bcd.index(str[i])
		if schar < 0 {
			return nil
		}
		if i%2 > 0 {
			buf[i/2] &= 0x0f
			buf[i/2] |= (uint8)(schar) << 4
		} else {
			buf[i/2] = 0xf0 | ((uint8)(schar) & 0x0f)
		}
	}
	return buf
}

func (bcd *Bcd) Decode(buf []byte) string {
	var strbytes []byte
	for i := 0; i < len(buf); i++ {
		var b byte
		b = bcd.byte(int(buf[i] & 0x0f))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)

		b = bcd.byte(int(buf[i] >> 4))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)
	}
	return string(strbytes)
}
