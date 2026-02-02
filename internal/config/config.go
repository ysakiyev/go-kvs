package config

type ServerConfig struct {
	NodeID        string   // "node-1", "node-2", etc.
	IsLeader      bool     // true for leader, false for followers
	Address       string   // "localhost:50051"
	FollowerAddrs []string // For leader: list of follower addresses
	LeaderAddr    string   // For followers: leader address
}
