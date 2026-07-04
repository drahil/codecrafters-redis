package configs

import (
	"flag"
	"strconv"
	"strings"
)

var ReplicaOf = flag.String("replicaof", "", "Master server address")
var Port = flag.Int("port", 6379, "TCP port to listen on")
var MasterHost string
var MasterPort int

func Init() {
	if *ReplicaOf != "" {
		parts := strings.Fields(*ReplicaOf)
		if len(parts) == 2 {
			MasterHost = parts[0]

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}

			MasterPort = port
		}
	}
}
