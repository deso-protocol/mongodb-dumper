module github.com/bitclout/deso-mongodb-dumper

go 1.16

replace github.com/bitclout/deso-core => ../core/

replace github.com/golang/glog => ../core/third_party/github.com/golang/glog

replace github.com/laser/go-merkle-tree => ../core/third_party/github.com/laser/go-merkle-tree

replace github.com/sasha-s/go-deadlock => ../core/third_party/github.com/sasha-s/go-deadlock

require (
	cloud.google.com/go/storage v1.15.0
	github.com/DataDog/datadog-go v4.5.0+incompatible
	github.com/bitclout/deso-core v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/badger/v3 v3.2103.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/fatih/structs v1.1.0
	github.com/go-delve/delve v1.5.0 // indirect
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529
	github.com/gorilla/mux v1.8.0
	github.com/h2non/bimg v1.1.5
	github.com/kevinburke/twilio-go v0.0.0-20210327194925-1623146bcf73
	github.com/laser/go-merkle-tree v0.0.0-20180821204614-16c2f6ea4444
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mmcloughlin/avo v0.0.0-20201105074841-5d2f697d268f // indirect
	github.com/nyaruka/phonenumbers v1.0.69
	github.com/pkg/errors v0.9.1
	github.com/rollbar/rollbar-go v1.4.0
	github.com/sasha-s/go-deadlock v0.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/twitchyliquid64/golang-asm v0.15.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0
	go.mongodb.org/mongo-driver v1.4.5
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	google.golang.org/api v0.46.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.29.0
)
