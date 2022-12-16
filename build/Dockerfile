FROM golang:1.17

LABEL ORGANIZATION="Photolab"
LABEL SERVICE="face_comparison"
LABEL MAINTAINERS="sizikov@pho.to"

ARG LDFLAGS
ARG CONFIG_FILE_NAME

ENV BIN_FILE /opt/face_comparison/app
ENV CODE_DIR /go/src/
ENV CONFIG_FILE /etc/face_comparison/${CONFIG_FILE_NAME}.toml

WORKDIR ${CODE_DIR}

#RUN apt-get update && apt-get install -y \
#    libmagickwand-dev \
#    libgraphicsmagick-dev \
#    imagemagick \
#    graphicsmagick \
#    && apt-get clean \
#    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
#    && pkg-config --cflags --libs MagickWand \
#    && export CGO_CFLAGS_ALLOW='-Xpreprocessor'

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . ${CODE_DIR}

RUN go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} ${CODE_DIR}/cmd/face_comparison/

COPY ./configs/${CONFIG_FILE_NAME}.toml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
