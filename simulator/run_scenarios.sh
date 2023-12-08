sudo modprobe ip6table_filter

sudo sysctl -w net.core.rmem_max=2500000
sudo sysctl -w net.core.wmem_max=2500000

count=0

cat scenarios.txt | while read scenario
do
    count=$((count+1))
    echo "Scenario $count: $scenario"

    cp ./logs/client/metrics_template.csv ./logs/client/metrics.csv

    docker-compose down
    # CLIENT="quic_impl" CLIENT_PARAMS="--parallel=20" SERVER="quic_impl" SERVER_PARAMS="--bytes=125000" SCENARIO="$scenario" docker-compose up
    CLIENT="quic_impl" CLIENT_PARAMS="--parallel=20" SERVER="quic_impl" SERVER_PARAMS="--bytes=250000" SCENARIO="$scenario" docker-compose up

    mv ./logs/client/metrics.csv ./logs/client/metrics-$count.csv
done

echo "Fim"