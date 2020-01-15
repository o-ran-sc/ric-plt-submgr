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
	"bytes"
)

const cLogBufferMaxSize = 1024
const cMsgBufferMaxSize = 2048

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type PduPackerIf interface {
	PduPack(logBuf []byte, data *PackedData) error
}

func PduPackerPack(entry PduPackerIf, trgBuf *PackedData) error {

	var logBuffer []byte = make([]byte, cLogBufferMaxSize)
	logBuffer[0] = 0

	if trgBuf != nil {
		trgBuf.Buf = make([]byte, cMsgBufferMaxSize)
	}
	err := entry.PduPack(logBuffer, trgBuf)
	if err == nil {
		return nil
	}
	return fmt.Errorf("Pack failed: err: %s, logbuffer: %s", err.Error(), logBuffer[:bytes.IndexByte(logBuffer, 0)])
}

func PduPackerPackAllocTrg(entry PduPackerIf, trgBuf *PackedData) (error, *PackedData) {
	dataPacked := trgBuf
	if dataPacked == nil {
		dataPacked = &PackedData{}
	}
	err := PduPackerPack(entry, dataPacked)
	if err != nil {
		return err, nil
	}
	return nil, dataPacked
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type PduUnPackerIf interface {
	PduUnPack(logBuf []byte, data *PackedData) error
}

func PduPackerUnPack(entry PduUnPackerIf, data *PackedData) error {
	var logBuffer []byte = make([]byte, cLogBufferMaxSize)

	logBuffer[0] = 0
	err := entry.PduUnPack(logBuffer, data)
	if err == nil {
		return nil
	}
	return fmt.Errorf("Unpack failed: err: %s, logbuffer: %s", err.Error(), logBuffer[:bytes.IndexByte(logBuffer, 0)])
}
