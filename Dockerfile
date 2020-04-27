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
FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:6-u18.04-nng as submgrcore

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump gdb valgrind

WORKDIR /tmp

#RUN git clone https://github.com/nokia/asn1c.git
#RUN cd asn1c && test -f configure || autoreconf -iv
#RUN cd asn1c &&  ./configure
#RUN cd asn1c && make
##RUN cd asn1c && make check
#RUN cd asn1c && make install

#
# Swagger
#
ARG SWAGGERVERSION=v0.19.0
ARG SWAGGERURL=https://github.com/go-swagger/go-swagger/releases/download/${SWAGGERVERSION}/swagger_linux_amd64
RUN wget --quiet ${SWAGGERURL} \
    && mv swagger_linux_amd64 swagger \
    && chmod +x swagger \
    && mv swagger /usr/local/bin/

#
# GO DELVE
#
RUN export GOBIN=/usr/local/bin/ ; \
    go get -u github.com/go-delve/delve/cmd/dlv \
    && go install github.com/go-delve/delve/cmd/dlv


#
# RMR
#
ARG RMRVERSION=3.6.5
ARG RMRLIBURL=https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr_${RMRVERSION}_amd64.deb/download.deb
ARG RMRDEVURL=https://packagecloud.io/o-ran-sc/staging/packages/debian/stretch/rmr-dev_${RMRVERSION}_amd64.deb/download.deb
RUN wget --content-disposition ${RMRLIBURL} && dpkg -i rmr_${RMRVERSION}_amd64.deb
RUN wget --content-disposition ${RMRDEVURL} && dpkg -i rmr-dev_${RMRVERSION}_amd64.deb
RUN rm -f rmr_${RMRVERSION}_amd64.deb rmr-dev_${RMRVERSION}_amd64.deb


RUN mkdir /manifests/
RUN echo "rmrlib ${RMRVERSION} ${RMRLIBURL}" >> /manifests/versions.txt
RUN echo "rmrdev ${RMRVERSION} ${RMRDEVURL}" >> /manifests/versions.txt
RUN echo "swagger ${SWAGGERVERSION} ${SWAGGERURL}" >> /manifests/versions.txt


WORKDIR /opt/submgr

###########################################################
#
###########################################################
FROM submgrcore as submgre2apbuild


ENV CFLAGS="-DASN_DISABLE_OER_SUPPORT"
ENV CGO_CFLAGS="-DASN_DISABLE_OER_SUPPORT"

COPY 3rdparty 3rdparty
RUN cd 3rdparty/E2AP-v01.00.00 && \
    gcc -c ${CFLAGS} -I. -g -fPIC *.c  && \
    gcc *.o -g -shared -o libe2ap.so && \
    cp libe2ap.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

RUN cd 3rdparty/E2SM-gNB-NRT_V4.0.1 && \
    gcc -c ${CFLAGS} -I. -g -fPIC *.c  && \
    gcc *.o -g -shared -o libgnbnrt.so && \
    cp libgnbnrt.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

RUN cd 3rdparty/E2SM-gNB-X2-V4.0.1 && \
    gcc -c ${CFLAGS} -I. -g -fPIC *.c  && \
    gcc *.o -g -shared -o libgnbx2.so && \
    cp libgnbx2.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

RUN echo "E2AP         E2AP-v01.00.00" >> /manifests/versions.txt
RUN echo "E2SM-gNB-NRT E2SM-gNB-NRT_V4.0.1" >> /manifests/versions.txt
RUN echo "E2SM-gNB-X2  E2SM-gNB-X2-V4.0.1" >> /manifests/versions.txt

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
RUN go mod tidy

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


RUN echo "rtmgrapi ${RTMGRVERSION} https://gerrit.o-ran-sc.org/r/ric-plt/rtmgr" >> /manifests/versions.txt

#
#
#
COPY pkg pkg
COPY cmd cmd

RUN mkdir -p /opt/bin && \
    go build -o /opt/bin/submgr cmd/submgr.go && \
    mkdir -p /opt/build/container/usr/local


RUN go mod tidy

RUN cp go.mod go.sum /manifests/
RUN grep gerrit /manifests/go.sum > /manifests/go_gerrit.sum


# unittest
COPY test/config-file.json test/config-file.json
ENV CFG_FILE=/opt/submgr/test/config-file.json
COPY test/uta_rtg.rt test/uta_rtg.rt
ENV RMR_SEED_RT=/opt/submgr/test/uta_rtg.rt 

#ENV CGO_LDFLAGS="-fsanitize=address"
#ENV CGO_CFLAGS="-fsanitize=address"

RUN go test -test.coverprofile /tmp/submgr_cover.out -count=1 -v ./pkg/control 
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

COPY --from=submgrbuild /manifests /manifests

COPY --from=submgrbuild /opt/bin/submgr /
COPY --from=submgrbuild /usr/local/include /usr/local/include
COPY --from=submgrbuild /usr/local/lib /usr/local/lib
RUN ldconfig

COPY run_submgr.sh /
RUN chmod 755 /run_submgr.sh

#default config
COPY config /opt/config
ENV CFG_FILE=/opt/config/submgr-config.yaml
ENV RMR_SEED_RT=/opt/config/submgr-uta-rtg.rt


ENTRYPOINT ["/submgr"]
