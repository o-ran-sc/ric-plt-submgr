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
#	Abstract:	Builds a container to compile Subscription Manager's code
#	Date:		28 May 2019
#
FROM golang:1.12  as submgrbuild

ENV HTTP_PROXY=http://10.144.1.10:8080
ENV HTTPS_PROXY=http://10.144.1.10:8080
RUN echo 'Acquire::http::Proxy "http://87.254.212.121:8080/";' > /etc/apt/apt.conf

COPY build/binary/compile.sh /
COPY . /opt/submgr

RUN apt-get update \
    && apt-get install -y git vim build-essential cmake ksh autotools-dev dh-autoreconf gawk autoconf-archive


RUN /compile.sh

FROM ubuntu

COPY --from=submgrbuild /opt/build/container/submgr /
COPY --from=submgrbuild /opt/build/container/submgr.yaml /
COPY build/container/run_submgr.sh /
COPY --from=submgrbuild /opt/build/container/usr/ /usr/
RUN ldconfig