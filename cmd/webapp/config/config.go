package config

import (
	"goTrackingUserLocation/internal/common"

	"github.com/spf13/viper"
)

var Registry *viper.Viper

func Set() {
	var err error
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	Registry = viper.GetViper()
	common.SERVER_PORT = Registry.GetString("SERVER_PORT")
	common.EMAIL_ADDRESS = Registry.GetString("EMAIL_ADDRESS")
	common.EMAIL_PORT = Registry.GetInt("EMAIL_PORT")
	common.EMAIL_APP_PASS = Registry.GetString("EMAIL_APP_PASS")
	common.EMAIL_HOST = Registry.GetString("EMAIL_HOST")
	common.EMAIL_SUBJECT = Registry.GetString("EMAIL_SUBJECT")
}
