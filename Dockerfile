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
FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:3-u18.04-nng as submgrprebuild

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump gdb

WORKDIR /tmp

ARG RMRVERSION=3.0.5
# Install RMr shared library
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr_${RMRVERSION}_amd64.deb && rm -rf rmr_${RMRVERSION}_amd64.deb
# Install RMr development header files
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr-dev_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr-dev_${RMRVERSION}_amd64.deb && rm -rf rmr-dev_${RMRVERSION}_amd64.deb

# "PULLING LOG and COMPILING LOG"
#RUN git clone "https://gerrit.o-ran-sc.org/r/com/log" /opt/log && cd /opt/log && \
# ./autogen.sh && ./configure && make install && ldconfig

# "Installing Swagger"
RUN cd /usr/local/go/bin \
    && wget --quiet https://github.com/go-swagger/go-swagger/releases/download/v0.19.0/swagger_linux_amd64 \
    && mv swagger_linux_amd64 swagger \
    && chmod +x swagger


WORKDIR /opt/submgr

RUN mkdir pkg

#
#
#
ENV CFLAGS="-DASN_DISABLE_OER_SUPPORT"
ENV CGO_CFLAGS="-DASN_DISABLE_OER_SUPPORT"

COPY 3rdparty 3rdparty
RUN cd 3rdparty/libe2ap && \
    gcc -c ${CFLAGS} -I. -fPIC *.c  && \
    gcc *.o -shared -o libe2ap.so && \
    cp libe2ap.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

COPY e2ap e2ap
RUN cd e2ap/libe2ap_wrapper && \
    gcc -c ${CFLAGS} -fPIC *.c  && \
    gcc *.o -shared -o libe2ap_wrapper.so && \
    cp libe2ap_wrapper.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

# unittest
RUN cd e2ap && /usr/local/go/bin/go test -v ./pkg/conv
RUN cd e2ap && /usr/local/go/bin/go test -v ./pkg/e2ap_wrapper

# test formating (not important)
RUN cd e2ap && test -z "$(/usr/local/go/bin/gofmt -l pkg/conv/*.go)"
RUN cd e2ap && test -z "$(/usr/local/go/bin/gofmt -l pkg/e2ap_wrapper/*.go)"
RUN cd e2ap && test -z "$(/usr/local/go/bin/gofmt -l pkg/e2ap/*.go)"
RUN cd e2ap && test -z "$(/usr/local/go/bin/gofmt -l pkg/e2ap/e2ap_tests/*.go)"


FROM submgrprebuild as submgrbuild
#
#
#
COPY go.mod go.mod
COPY go.sum go.sum

RUN /usr/local/go/bin/go mod download

#
#
#
COPY api api

# "Getting and generating routing managers api client"
RUN git clone "https://gerrit.o-ran-sc.org/r/ric-plt/rtmgr" \
    && cp rtmgr/api/routing_manager.yaml api/ \
    && rm -rf rtmgr

RUN mkdir -p /root/go && \
    /usr/local/go/bin/swagger generate client -f api/routing_manager.yaml -t pkg/ -m rtmgr_models -c rtmgr_client

#
#
#
COPY pkg pkg
COPY cmd cmd

RUN mkdir -p /opt/bin && \
    /usr/local/go/bin/go build -o /opt/bin/submgr cmd/submgr.go && \
    mkdir -p /opt/build/container/usr/local


RUN /usr/local/go/bin/go mod tidy

# unittest
COPY test/config-file.json test/config-file.json
ENV CFG_FILE=/opt/submgr/test/config-file.json

RUN /usr/local/go/bin/go test -test.coverprofile /tmp/submgr_cover.out -count=1 -v ./pkg/control

RUN /usr/local/go/bin/go tool cover -html=/tmp/submgr_cover.out -o /tmp/submgr_cover.html

# test formating (not important)
RUN test -z "$(/usr/local/go/bin/gofmt -l pkg/control/*.go)"

#
#
#
FROM ubuntu:18.04

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump

COPY run_submgr.sh /
COPY --from=submgrbuild /opt/bin/submgr /
COPY --from=submgrbuild /usr/local/include /usr/local/include
COPY --from=submgrbuild /usr/local/lib /usr/local/lib
RUN ldconfig

RUN chmod 755 /run_submgr.sh
CMD /run_submgr.sh
