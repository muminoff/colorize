FROM tensorflow/tensorflow

RUN curl -SL https://s3.amazonaws.com/tinyclouds-storage/colorize-20160108.tgz \
	| tar -xzC / \
	&& mv /colorize-20160108 /colorize

RUN pip install scikit-image

RUN curl -SL https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz \
	| tar -xzC /usr/local

ENV GOPATH /go
ENV PATH $PATH:/usr/local/go/bin:$GOPATH/bin

COPY colorize.py /colorize/

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY . /go/src/app
RUN go get

ENTRYPOINT ["/go/bin/app"]
