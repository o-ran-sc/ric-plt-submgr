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
FROM nexus3.o-ran-sc.org:10004/bldr-ubuntu18-c-go:1-u18.04-nng1.1.1 as submgrbuild

COPY . /opt/submgr

# Install RMr shared library
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/master/packages/debian/stretch/rmr_1.0.36_amd64.deb/download.deb && dpkg -i rmr_1.0.36_amd64.deb
# Install RMr development header files
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/master/packages/debian/stretch/rmr-dev_1.0.36_amd64.deb/download.deb && dpkg -i rmr-dev_1.0.36_amd64.deb

# "PULLING LOG and COMPILING LOG"
RUN git clone "https://gerrit.o-ran-sc.org/r/com/log" /opt/log && cd /opt/log && \
 ./autogen.sh && ./configure && make install && ldconfig &&
  mkdir -p /opt/bin && cd /opt/submgr && \
   /usr/local/go/bin/go get && \
    /usr/local/go/bin/go build -o /opt/bin/submgr ./cmd/submgr.go && \
     /usr/local/go/bin/go build -o /opt/test/rco/rco ./test/rco/rco.go && \
      /usr/local/go/bin/go build -o /opt/test/e2t/e2t ./test/e2t/e2t.go && \
       mkdir -p /opt/build/container/usr/local /opt/test/e2t/container/usr/local /opt/test/rco/container/usr/local && \
        cp -Rf /usr/local/lib /usr/local/include /opt/build/container/usr/local/ && \
         cp -Rf /usr/local/lib /usr/local/include  /opt/test/e2t/container/usr/local/ && \
          cp -Rf /usr/local/lib /usr/local/include  /opt/test/rco/container/usr/local/ && \
           cp bin/submgr config/submgr.yaml /opt/build/container/  && \
            rm -Rf /opt/rmr /opt/log /opt/go.sum

FROM ubuntu:18.04

COPY --from=submgrbuild /opt/build/container/submgr /
COPY --from=submgrbuild /opt/build/container/submgr.yaml /
COPY run_submgr.sh /
COPY --from=submgrbuild /opt/build/container/usr/ /usr/
RUN ldconfig
CMD /run_submgr.sh