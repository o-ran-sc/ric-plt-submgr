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
package teststub

import (
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"io/ioutil"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestError(t *testing.T, pattern string, args ...interface{}) {
	xapp.Logger.Error(fmt.Sprintf(pattern, args...))
	t.Errorf(fmt.Sprintf(pattern, args...))
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestLog(t *testing.T, pattern string, args ...interface{}) {
	xapp.Logger.Info(fmt.Sprintf(pattern, args...))
	t.Logf(fmt.Sprintf(pattern, args...))
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateTmpFile(str string) (string, error) {
	file, err := ioutil.TempFile("/tmp", "*.rt")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(str)
	if err != nil {
		file.Close()
		return "", err
	}
	return file.Name(), nil
}
