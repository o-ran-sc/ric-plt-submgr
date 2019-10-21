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
#	Abstract:	Builds RCO stub container
#  	Date:		28 May 2019
#

FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:1-u18.04-nng1.1.1 as submgrbuild

ENV HTTP_PROXY=http://10.144.1.10:8080
ENV HTTPS_PROXY=http://10.144.1.10:8080
RUN echo 'use_proxy=yes \n http_proxy=http://10.144.1.10:8080 \n https_proxy=http://10.144.1.10:8080' > /etc/wgetrc
RUN echo 'Acquire::http::Proxy "http://87.254.212.121:8080/";' > /etc/apt/apt.conf

# Install RMr shared library
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr_1.6.0_amd64.deb/download.deb && dpkg -i rmr_1.6.0_amd64.deb
# Install RMr development header files
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr-dev_1.6.0_amd64.deb/download.deb && dpkg -i rmr-dev_1.6.0_amd64.deb

# "PULLING LOG and COMPILING LOG"
RUN git clone "https://gerrit.o-ran-sc.org/r/com/log" /opt/log && cd /opt/log && \
 ./autogen.sh && ./configure && make install && ldconfig

WORKDIR /opt/submgr

# "COMPILING E2AP Wrapper"
COPY e2ap e2ap
RUN cd e2ap && \
 gcc -c -fPIC -Iheaders/ lib/*.c wrapper.c && \
 gcc *.o -shared -o libwrapper.so && \
 cp libwrapper.so /usr/local/lib/ && \
 cp wrapper.h headers/*.h /usr/local/include/ && \
 ldconfig

# "Installing Swagger"
RUN cd /usr/local/go/bin \
    && wget --quiet https://github.com/go-swagger/go-swagger/releases/download/v0.19.0/swagger_linux_amd64 \
    && mv swagger_linux_amd64 swagger \
    && chmod +x swagger

# "Getting and generating routing managers api client"
COPY api api
RUN git clone "https://gerrit.o-ran-sc.org/r/ric-plt/rtmgr" \
    && cp rtmgr/api/routing_manager.yaml api/ \
    && rm -rf rtmgr

RUN mkdir pkg

COPY go.mod go.mod

RUN mkdir -p /root/go && \
    /usr/local/go/bin/swagger generate client -f api/routing_manager.yaml -t pkg/ -m rtmgr_models -c rtmgr_client

COPY go.mod go.mod

RUN /usr/local/go/bin/go mod tidy

#Compiling RCO
COPY pkg pkg
COPY cmd cmd
COPY test test

RUN mkdir -p /opt/bin && \
   /usr/local/go/bin/go build -o /opt/test/rco/rco ./test/rco/rco.go && \
    mkdir -p /opt/test/rco/container/usr/local

FROM ubuntu:18.04

COPY --from=submgrbuild /opt/test/rco/rco /
COPY  test/rco/rco.yaml test/rco/container/run_rco.sh test/uta_rtg.rt /
COPY --from=submgrbuild /usr/local/include /usr/local/include
COPY --from=submgrbuild /usr/local/lib /usr/local/lib

RUN ldconfig
RUN chmod 755 /run_rco.sh
RUN chmod 755 /rco
