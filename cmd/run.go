package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"

	coreCmd "github.com/bitclout/core/cmd"
	"github.com/golang/glog"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the MongoDB dumper",
	Long: `...`,
	Run: Run,
}

func Run(cmd *cobra.Command, args []string) {
	// Start the core node
	coreConfig := coreCmd.LoadConfig()
	coreNode := coreCmd.NewNode(coreConfig)
	coreNode.Start()

	// Start the mongo dumper
	mongoConfig := LoadConfig()
	mongoNode := NewNode(mongoConfig, coreNode)
	mongoNode.Start()

	shutdownListener := make(chan os.Signal)
	signal.Notify(shutdownListener, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		coreNode.Stop()
		mongoNode.Stop()
		glog.Info("Shutdown complete")
	}()

	<-shutdownListener
}

func init() {
	// Add all the core node flags
	coreCmd.SetupRunFlags(runCmd)

	// Add the mongo dumper flags
	runCmd.PersistentFlags().String("mongo-uri", "mongodb://localhost:27017", "Mongo connection URI")
	runCmd.PersistentFlags().String("mongo-database", "bitclout", "Mongo database name")
	runCmd.PersistentFlags().String("mongo-collection", "data", "Mongo collection name")

	runCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		viper.BindPFlag(flag.Name, flag)
	})

	rootCmd.AddCommand(runCmd)
}
