docker run -d -p 2379:2379  -p 2380:2380
--mount type=bind,source=D:/app/etcd/etcd-data.tmp,destination=/etcd-data
--network etcdnet
--name etcd-gcr-v3.4.15 quay.io/coreos/etcd:v3.4.15 /usr/local/bin/etcd
--name s1 --data-dir /etcd-data
--listen-client-urls http://0.0.0.0:2379
--advertise-client-urls http://0.0.0.0:2379
--listen-peer-urls http://0.0.0.0:2380
--initial-advertise-peer-urls http://0.0.0.0:2380
--initial-cluster s1=http://0.0.0.0:2380
--initial-cluster-token tkn
--initial-cluster-state new --log-level info
--logger zap --log-outputs stderr
curl -i -X GET http://localhost:8888/api/order/get/1