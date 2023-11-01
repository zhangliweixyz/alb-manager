package conf

import "github.com/spf13/viper"

var ViperConfig *viper.Viper

// 全局配置读取
func InitConfig() error {
	ViperConfig = viper.New()
	ViperConfig.SetConfigFile("./conf/config.yaml")
	if err := ViperConfig.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
