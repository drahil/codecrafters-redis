package configs

import "flag"

var ReplicaOf = flag.String("replicaof", "", "Master server address")
var Port = flag.Int("port", 6379, "TCP port to listen on")
