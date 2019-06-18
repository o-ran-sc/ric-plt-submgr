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
#	Mnemonic:	build.sh
#	Abstract:	Compiles the Subscription Manager's source and builds the docker container
#	Date:		28 May 2019
#

echo 'Creating compiler container'
docker build --no-cache --tag=submgr_compiler:0.1 $PWD/build/binary/

echo 'Running submgr compiler'
docker run --rm -v ${PWD}:/opt submgr_compiler:0.1

echo 'Cleaning up compiler container'
docker rmi -f submgr_compiler:0.1

echo 'submgr binary successfully built!'

echo 'Creating submgr container'
docker build --no-cache --tag=submgr:builder ${PWD}/build/container/

