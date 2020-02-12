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
###########################################################
#
###########################################################
FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:4-u18.04-nng as submgrcore

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump gdb valgrind

WORKDIR /tmp

ARG RMRVERSION=1.13.1
# Install RMr shared library
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr_${RMRVERSION}_amd64.deb && rm -rf rmr_${RMRVERSION}_amd64.deb
# Install RMr development header files
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr-dev_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr-dev_${RMRVERSION}_amd64.deb && rm -rf rmr-dev_${RMRVERSION}_amd64.deb

# "Installing Swagger"
RUN wget --quiet https://github.com/go-swagger/go-swagger/releases/download/v0.19.0/swagger_linux_amd64 \
    && mv swagger_linux_amd64 swagger \
    && chmod +x swagger \
    && mkdir -p /root/.go/bin \
    && mv swagger /root/.go/bin

ENV GOPATH=/root/.go
ENV PATH=$PATH:/root/.go/bin
RUN go get -u github.com/go-delve/delve/cmd/dlv

WORKDIR /opt/submgr

###########################################################
#
###########################################################
FROM submgrcore as submgre2apbuild


ENV CFLAGS="-DASN_DISABLE_OER_SUPPORT"
ENV CGO_CFLAGS="-DASN_DISABLE_OER_SUPPORT"

COPY 3rdparty 3rdparty
RUN cd 3rdparty/libe2ap && \
    gcc -c ${CFLAGS} -I. -g -fPIC *.c  && \
    gcc *.o -g -shared -o libe2ap.so && \
    cp libe2ap.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

COPY e2ap e2ap
RUN cd e2ap/libe2ap_wrapper && \
    gcc -c ${CFLAGS} -g -fPIC *.c  && \
    gcc *.o -g -shared -o libe2ap_wrapper.so && \
    cp libe2ap_wrapper.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

# unittest
RUN cd e2ap && go test -v ./pkg/conv
RUN cd e2ap && go test -v ./pkg/e2ap_wrapper

# test formating (not important)
RUN cd e2ap && test -z "$(gofmt -l pkg/conv/*.go)"
RUN cd e2ap && test -z "$(gofmt -l pkg/e2ap_wrapper/*.go)"
RUN cd e2ap && test -z "$(gofmt -l pkg/e2ap/*.go)"
RUN cd e2ap && test -z "$(gofmt -l pkg/e2ap/e2ap_tests/*.go)"


###########################################################
#
###########################################################
FROM submgre2apbuild as submgrbuild
#
#
#
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

#
#
#
RUN mkdir pkg
COPY api api


ARG RTMGRVERSION=cd7867c8f527f46fd8702b0b8d6b380a8e134bea

RUN git clone "https://gerrit.o-ran-sc.org/r/ric-plt/rtmgr" \
    && git -C "rtmgr" checkout $RTMGRVERSION \
    && cp rtmgr/api/routing_manager.yaml api/ \
    && rm -rf rtmgr


RUN mkdir -p /root/go && \
    swagger generate client -f api/routing_manager.yaml -t pkg/ -m rtmgr_models -c rtmgr_client

#
#
#
COPY pkg pkg
COPY cmd cmd

RUN mkdir -p /opt/bin && \
    go build -o /opt/bin/submgr cmd/submgr.go && \
    mkdir -p /opt/build/container/usr/local


RUN go mod tidy

# unittest
COPY test/config-file.json test/config-file.json
ENV CFG_FILE=/opt/submgr/test/config-file.json

RUN go test -test.coverprofile /tmp/submgr_cover.out -count=1 -v ./pkg/control 

#-c -o submgr_test
#RUN ./submgr_test -test.coverprofile /tmp/submgr_cover.out

RUN go tool cover -html=/tmp/submgr_cover.out -o /tmp/submgr_cover.html

# test formating (not important)
RUN test -z "$(gofmt -l pkg/control/*.go)"
RUN test -z "$(gofmt -l pkg/teststub/*.go)"
RUN test -z "$(gofmt -l pkg/teststubdummy/*.go)"
RUN test -z "$(gofmt -l pkg/teststube2ap/*.go)"
RUN test -z "$(gofmt -l pkg/xapptweaks/*.go)"


###########################################################
#
###########################################################
FROM ubuntu:18.04

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump

COPY run_submgr.sh /
COPY --from=submgrbuild /opt/bin/submgr /
COPY --from=submgrbuild /usr/local/include /usr/local/include
COPY --from=submgrbuild /usr/local/lib /usr/local/lib
RUN ldconfig

RUN chmod 755 /run_submgr.sh
CMD /run_submgr.sh
