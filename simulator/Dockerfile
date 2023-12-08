FROM martenseemann/quic-network-simulator-endpoint:latest

# Tutorial for use with QUIC Network Simulator: https://github.com/marten-seemann/quic-network-simulator

# Erro no serviço "sim":ip6tables v1.8.4 (legacy): can't initialize ip6tables table `filter': Table does not exist (do you need to insmod?)
# Solução: https://ilhicas.com/2018/04/08/Fixing-do-you-need-insmod.html (sudo modprobe ip6table_filter)

RUN apt-get update

# Download and build your QUIC implementation
RUN apt-get install -y wget tar git vim python

RUN wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz && \
  tar xfz go1.21.4.linux-amd64.tar.gz && \
  rm go1.21.4.linux-amd64.tar.gz

ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on

# RUN git clone https://github.com/eriklima/http3-quic.git

COPY http3-quic/ http3-quic/

WORKDIR /http3-quic
# RUN go mod tidy

# Build HTTP3's client and server
# RUN go build -o client/client ./client/client.go
# RUN go build -o server/server ./server/server.go

# Copy run script and run it
COPY run_endpoint.sh .
RUN chmod +x run_endpoint.sh
ENTRYPOINT [ "./run_endpoint.sh" ]