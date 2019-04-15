package cmd

import (
	"log"
	"os"

	"github.com/kzap/ada-backend-challenge/internal/config"
	"github.com/kzap/ada-backend-challenge/internal/webserver"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "backend-challenge",
	Short: "Ada Backend Challenge",
	Long: `Ada Backend Challenge
			https://github.com/AdaSupport/backend-challenge`,
	Version: "0.1",
	Run: func(cmd *cobra.Command, args []string) {
		MySQLConfig.Host = viper.GetString("db_host")
		MySQLConfig.Port = viper.GetInt("db_port")
		MySQLConfig.DBName = viper.GetString("db_name")
		MySQLConfig.Username = viper.GetString("db_username")
		MySQLConfig.Password = viper.GetString("db_password")

		webserver.Start(MySQLConfig)
	},
}

var MySQLConfig config.DbConfig

func init() {
	viper.SetEnvPrefix("ada")
	viper.AutomaticEnv()
	flags := rootCmd.PersistentFlags()
	flags.StringVar(&MySQLConfig.Host, "db_host", "localhost", "mysql database host | env var: [ADA_DB_HOST]")
	flags.IntVar(&MySQLConfig.Port, "db_port", 3306, "mysql database port | env var: [ADA_DB_PORT]")
	flags.StringVar(&MySQLConfig.DBName, "db_name", "ada_test", "mysql database name | env var: [ADA_DB_NAME]")
	flags.StringVar(&MySQLConfig.Username, "db_username", "ada_test", "mysql database user | env var: [ADA_DB_USERNAME]")
	flags.StringVar(&MySQLConfig.Password, "db_password", "ada_test", "mysql database password | env var: [ADA_DB_PASSWORD]")
	viper.BindPFlags(flags)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
