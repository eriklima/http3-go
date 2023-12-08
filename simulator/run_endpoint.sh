#!/bin/bash

# Set up the routing needed for the simulation
/setup.sh

# The following variables are available for use:
# - ROLE contains the role of this execution context, client or server
# - SERVER_PARAMS contains user-supplied command line parameters
# - CLIENT_PARAMS contains user-supplied command line parameters

# echo "193.167.0.100 client" >> /etc/hosts
# echo "193.167.100.100 server" >> /etc/hosts

if [ "$ROLE" == "client" ]; then
    echo "Wait for the simulator to start up."
    /wait-for-it.sh sim:57832 -s -t 30
    /wait-for-it.sh server:4433 -s -t 5

    echo "Request server HTTP3 to 193.167.100.100:4433"
    echo "Client params: $CLIENT_PARAMS"

    # ./client/client --server=193.167.100.100:4433 $CLIENT_PARAMS

    cd client/

    for exper in {1..4};do
        echo "Experimento (payload): $exper";
        for i in {1..11};do
            echo "Repetição: $i"
            ./client --server=193.167.100.100:4433 $CLIENT_PARAMS -expernumber=$exper
        done
    done

    # stoppar o server (SIGINT) kill -2 [PID]
    # stoppar o simulador (sim)

    # echo "Termina processo Server"
    # ps aux | grep "addr=0.0.0.0" | awk '{print $2}' | xargs kill -2

    # echo "Termina processo Simulador"
    # ps aux | grep "simple-p2p" | awk '{print $2}' | xargs kill -2

elif [ "$ROLE" == "server" ]; then
    # It is recommended to increase the maximum buffer size (https://github.com/quic-go/quic-go/wiki/UDP-Receive-Buffer-Size)
    # sysctl -w net.core.rmem_max=2500000

    echo "Run the server HTTP3 on 0.0.0.0:4433"
    echo "Server params: $SERVER_PARAMS"

    # ./server/server --addr=0.0.0.0:4433 $SERVER_PARAMS

    cd server/
    ./server --addr=0.0.0.0:4433 $SERVER_PARAMS
fi