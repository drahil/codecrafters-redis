package command

type ClientState struct {
	InMulti   bool
	IsReplica bool
	Queue     [][]string
}

func (client *ClientState) InitializeMulti() {
	client.InMulti = true
	client.Queue = nil
}
