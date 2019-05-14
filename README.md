# Cluster
defines the communication strategy between servers, and implments message router&forwarding, auto service discovery and load balance.

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
