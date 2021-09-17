package cmd

import (
	coreCmd "github.com/bitclout/deso-core/cmd"
	"github.com/bitclout/deso-mongodb-dumper/mongodb"
)

type Node struct {
	SyncingService *mongodb.SyncingService
	Config         *Config

	CoreNode       *coreCmd.Node
}

func NewNode(config *Config, coreNode *coreCmd.Node) *Node {
	result := Node{}
	result.Config = config
	result.CoreNode = coreNode

	return &result
}

func (node *Node) Start() {
	node.SyncingService = mongodb.NewSyncingService(
		node.CoreNode.Server.GetBlockchain().DB(),
		node.Config.MongoURI,
		node.Config.MongoDatabase,
		node.Config.MongoCollection)

	go func() {
		node.SyncingService.ConnectToMongo()
		node.SyncingService.Start()
	}()
}

func (node *Node) Stop() {
	node.SyncingService.DisconnectFromMongo()
}
