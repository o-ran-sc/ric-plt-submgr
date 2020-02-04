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

type PduPackerIf interface {
	PduPack(logBuf []byte, data *PackedData) error
}

func PduPackerPack(entry PduPackerIf) (error, *PackedData) {
	var logBuffer []byte = make([]byte, cLogBufferMaxSize)
	logBuffer[0] = 0

	trgBuf := &PackedData{}
	err := entry.PduPack(logBuffer, trgBuf)
	if err == nil {
		return nil, trgBuf
	}
	return fmt.Errorf("Pack failed: err: %s, logbuffer: %s", err.Error(), logBuffer[:strings.Index(string(logBuffer[:]), "\000")]), nil
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
	var logBuffer []byte = make([]byte, cLogBufferMaxSize)
	logBuffer[0] = 0
	err := entry.PduUnPack(logBuffer, data)
	if err == nil {
		return nil
	}
	return fmt.Errorf("Unpack failed: err: %s, logbuffer: %s", err.Error(), logBuffer[:strings.Index(string(logBuffer[:]), "\000")])
}
