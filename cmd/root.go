/*
Copyright © 2020 David Muckle <dvdmuckle@dvdmuckle.xyz>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/dvdmuckle/goify/cmd/helper"
	"github.com/golang/glog"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//Config type stores constantly retrieved things from the config file

var conf helper.Config

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goify",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.config/goify/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := home + "/.config/goify"
		if err := os.MkdirAll(configPath, 0755); err != nil {
			glog.Fatal("Error creating config path: ", err)
		}

		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		cfgFile = fmt.Sprintf(configPath + "/config.yaml")
	}
	viper.SetDefault("spotifyclientid", "")
	viper.SetDefault("spotifysecret", "")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	if err := viper.WriteConfigAs(cfgFile); err != nil {
		glog.Fatal("Error writing config file: ", err)
	}
	conf.ClientID = viper.GetString("spotifyclientid")
	if secret, err := base64.StdEncoding.DecodeString(viper.GetString("spotifysecret")); err != nil && len(secret) != 0 {
		//Do nothing
		fmt.Println("Hit error on decoding secert")
	} else {
		conf.Secret = strings.TrimSpace(string(secret))
	}
	//TODO: #2 I would love to do something like viper.GetStringMapString("auth") here but
	//I can't figure out how to cast those results to type oauth2.Token
	conf.Token.AccessToken = viper.GetString("auth.accesstoken")
	conf.Token.RefreshToken = viper.GetString("auth.refreshtoken")
	conf.Token.TokenType = viper.GetString("auth.tokentype")
	conf.Token.Expiry = viper.GetTime("auth.expiry")
}
