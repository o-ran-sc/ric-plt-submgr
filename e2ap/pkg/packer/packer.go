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

package packer

import (
	"fmt"
	"strings"
)

const cLogBufferMaxSize = 40960

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type PduLoggerBuf struct {
	logBuffer []byte
}

func (lb *PduLoggerBuf) String() string {
	return "logbuffer(" + string(lb.logBuffer[:strings.Index(string(lb.logBuffer[:]), "\000")]) + ")"
}

func NewPduLoggerBuf() *PduLoggerBuf {
	lb := &PduLoggerBuf{}
	lb.logBuffer = make([]byte, cLogBufferMaxSize)
	lb.logBuffer[0] = 0
	return lb
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type PduPackerIf interface {
	PduPack(logBuf []byte) (error, *PackedData)
}

func PduPackerPack(entry PduPackerIf) (error, *PackedData) {
	lb := NewPduLoggerBuf()
	err, buf := entry.PduPack(lb.logBuffer)
	if err == nil {
		return nil, buf
	}
	return fmt.Errorf("Pack failed: err(%s), %s", err.Error(), lb.String()), nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type PduUnPackerIf interface {
	PduUnPack(logBuf []byte, data *PackedData) error
}

func PduPackerUnPack(entry PduUnPackerIf, data *PackedData) error {
	if data == nil {
		return fmt.Errorf("Unpack failed: data is nil")
	}
	lb := NewPduLoggerBuf()
	err := entry.PduUnPack(lb.logBuffer, data)
	if err == nil {
		return nil
	}
	return fmt.Errorf("Unpack failed: err(%s), %s", err.Error(), lb.String())
}
