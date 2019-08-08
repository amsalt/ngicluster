# Cluster
defines the communication strategy between servers, and implments message router&relay, auto service discovery and load balance.

## featrues
- load balance
- service discovery
- message forwarding
- configurable heartbeat check

## balancer
- loader balancer
- random balancer
- round-robin balancer
- stickiness balancer

## resolver
- zookeeper resolver
- etcd resolver
- static config resovler

## relay & router
defines message id in router, and will be auto relay to the correct receiver.

## structure
```
.
├── README.md
├── balancer    负载均衡器。
│   ├── balancer.go
│   ├── builder.go
│   ├── loader  按负载做均衡，每次选择负载最小的节点服务。
│   │   └── loader.go
│   ├── random  随机负载均衡，适用于短时间内高并发的场景。
│   │   └── random.go
│   ├── roundrobin  roundrobin负载均衡算法。
│   │   └── roundrobin.go
│   └── stickiness  有粘性的负载均衡器。
│       └── stickiness.go
├── client.go       rpc客户端。
├── client_mgr.go   rpc客户端管理器。
├── cluster.go
├── consts
│   └── consts.go
├── extra_msg.go
├── handler.go
├── heartbeat_handler.go
├── relay_handler.go
├── relay_router.go
├── resolver    服务发现。
│   ├── etcd    基于etcd的服务发现服务。
│   │   └── etcd.go
│   ├── resolver.go
│   ├── static  基于配置的服务发现服务。
│   │   └── static.go
│   └── zookeeper   基于zookeeper的服务发现服务。
│       └── zookeeper.go
├── server.go       rpc服务器端。
└── test    测试用例。
    ├── balancer_test.go
    ├── cluster_test.go
    └── init.go
```
