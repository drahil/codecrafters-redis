package configs

import "flag"

var ReplicaOf = flag.String("replicaof", "", "Master server address")

var Config = map[string]string{}

func Init() {
	
}