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

package control

import (
  "encoding/gob"
  "bytes"
  "errors"
)

type E2ap struct {
}

func (c *E2ap) GetSubscriptionSequenceNumber(payload []byte) (int, error) {
  asn1 := new(Asn1)
  message, err := asn1.Decode(payload)
  if err != nil {
    return 0, errors.New("Unable to decode payload due to "+ err.Error()) 
  }
  return message.SubscriptionId, nil
}

func (c *E2ap) SetSubscriptionSequenceNumber(payload []byte, newSubscriptionid int) ([]byte ,error) {
  asn1 := new(Asn1)
  message, err := asn1.Decode(payload)
  if err != nil {
    return make([]byte,0), errors.New("Unable to decode payload due to "+ err.Error()) 
  }
  message.SubscriptionId = newSubscriptionid
  payload, err = asn1.Encode(message)
  if err != nil {
    return make([]byte,0), errors.New("Unable to encode message due to "+ err.Error()) 
  }
  return payload, nil
}


func (c *E2ap) GetPayloadContent(payload []byte) (content string, err error) {
  asn1 := new(Asn1)
  message, err := asn1.Decode(payload)
  content = message.Content
  return
}
/*
Serialize and Deserialize message using this until real ASN1 GO wrapper is not in place
*/
type Asn1 struct {
}

func (a *Asn1) Encode(message RmrPayload) ([]byte, error) {
	buffer := new(bytes.Buffer)
	asn1 := gob.NewEncoder(buffer)
	if err := asn1.Encode(message); err != nil {
		return nil, err
  }
	return buffer.Bytes(), nil
}

func (a *Asn1) Decode(data []byte) (RmrPayload, error) {
  message := new(RmrPayload)
  buffer := bytes.NewBuffer(data)
	asn1 := gob.NewDecoder(buffer)
	if err := asn1.Decode(message); err != nil {
		return RmrPayload{}, err
  }
	return *message, nil
}
