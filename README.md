# http3-quic

Implementação de um cliente e um servidor HTTP3 sobre a camada de transporte QUIC.

## Como Utilizar o QUIC Network Simulator

Erro no serviço "sim":ip6tables v1.8.4 (legacy): can't initialize ip6tables table `filter': Table does not exist (do you need to insmod?)

Solução: https://ilhicas.com/2018/04/08/Fixing-do-you-need-insmod.html

`sudo modprobe ip6table_filter`

### Comandos para aumentar o tamanho do buffer UDP

`sudo sysctl -w net.core.rmem_max=2500000`

`sudo sysctl -w net.core.wmem_max=2500000`

### Comando para executar fazer o build do docker-compose:

`CLIENT="quic_impl" SERVER="quic_impl" docker-compose build [--no-cache] [--parallel]`

### Comando para executar o docker-compose:

`CLIENT="quic_impl" SERVER="quic_impl" SCENARIO="simple-p2p --delay=15ms --bandwidth=10Mbps --queue=25" docker-compose up`

### Comando para executar o docker-compose com parâmetros:

`CLIENT="quic_impl" 
CLIENT_PARAMS="-qlog -qlogpath="/logs/qlog" -parallel=10 -bytes=10000000" 
SERVER="quic_impl" 
SERVER_PARAMS="-qlog -qlogpath="/logs/qlog"" 
SCENARIO="simple-p2p --delay=15ms --bandwidth=10Mbps --queue=25" 
docker-compose up`

## Visualisar logs QLog

1. Acessar: https://qvis.quictools.info/#/files
2. Importar os arquivos QLog do _client_ e _server_
