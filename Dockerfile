FROM golang:1.11 as builder

RUN go get -u -v github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/refunc/aws-api-gw

COPY . /go/src/github.com/refunc/aws-api-gw/

RUN dep ensure -v \
    && CGO_ENABLED=0 go build -tags netgo -installsuffix netgo -ldflags "-s -w" -o /invoke main.go

FROM python:3.7

COPY requirements.txt /opt/aws-api-gw/

WORKDIR /opt/aws-api-gw

RUN export PIP_DOWNLOAD_CACHE=/tmp \
    && pip install -r requirements.txt  \
    && rm -rf /tmp/* && rm -rf ~/.cache

COPY --from=builder /invoke /bin/

COPY . ${WORKDIR}

CMD python main.py
