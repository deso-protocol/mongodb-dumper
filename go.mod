module github.com/deso-protocol/mongodb-dumper

go 1.16

replace github.com/deso-protocol/core => ../core/

require (
	github.com/deso-protocol/core v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/badger/v3 v3.2103.0
	github.com/fatih/structs v1.1.0
	github.com/golang/glog v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	go.mongodb.org/mongo-driver v1.4.5
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
