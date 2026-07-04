package configs

import (
	"flag"
	"strconv"
	"strings"
)

type Configs struct {
	Port       int
	ReplicaOf  string
	MasterHost string
	MasterPort int
}

func Init() Configs {
	port := flag.Int("port", 6379, "TCP port to listen on")
	replicaOf := flag.String("replicaof", "", "Master server address")

	flag.Parse()

	masterHost := ""
	masterPort := -1

	if *replicaOf != "" {
		parts := strings.Fields(*replicaOf)
		if len(parts) == 2 {
			masterHost = parts[0]

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}

			masterPort = port
		}
	}

	return Configs{
		Port:       *port,
		ReplicaOf:  *replicaOf,
		MasterHost: masterHost,
		MasterPort: masterPort,
	}
}
