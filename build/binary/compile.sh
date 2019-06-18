#!/bin/sh -e
#
#==================================================================================
#   Copyright (c) 2019 AT&T Intellectual Property.
#   Copyright (c) 2019 Nokia
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#==================================================================================
#
#
#	Mnemonic:	compile.sh
#	Abstract:	Compiles the source of Subscription Manager service and the two platform component stubs
#	Date:		28 May 2019
#
echo "PULLING RMR"
cd /opt
git clone "https://gerrit.o-ran-sc.org/r/ric-plt/lib/rmr"
echo "START COMPILING RMR"
cd /opt/rmr
git checkout v1.0.31
mkdir -p build
cd build
cmake ..
make install

echo "PULLING LOG"
cd /opt
git clone "https://gerrit.o-ran-sc.org/r/com/log"
echo "START COMPILING LOG"
cd /opt/log
./autogen.sh
./configure
make install

ldconfig

echo "DOWNLOAD GO DEPENDENCIES"
cd /opt/
mkdir -p /opt/bin
go get

echo "START COMPILING COMPONENTS"
go build -o /opt/bin/submgr ./cmd/submgr.go
go build -o /opt/test/rco/rco ./test/rco/rco.go
go build -o /opt/test/e2t/e2t ./test/e2t/e2t.go

echo "SAVE RESULT"
mkdir -p /opt/build/container/usr/local /opt/test/e2t/container/usr/local /opt/test/rco/container/usr/local
cp -Rf /usr/local/lib /usr/local/include /opt/build/container/usr/local/
cp -Rf /usr/local/lib /usr/local/include  /opt/test/e2t/container/usr/local/
cp -Rf /usr/local/lib /usr/local/include  /opt/test/rco/container/usr/local/
cp bin/submgr config/submgr.yaml /opt/build/container/

echo "CLEANUP"
rm -Rf /opt/rmr /opt/log /opt/go.sum

echo "DONE"
