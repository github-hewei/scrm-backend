package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var v = viper.New()

// Init 初始化配置（只读一次文件到内存）。
func Init() {
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] .env file not found\n")
	}

	if err := v.ReadInConfig(); err != nil {
		log.Printf("[WARN] read config.yaml failed: %v\n", err)
	}

	// 将 env 覆盖后的值固化到 Viper 内部，解决 UnmarshalKey 不合并 env 的问题
	for _, key := range v.AllKeys() {
		v.Set(key, v.Get(key))
	}
}

// UnmarshalKey 从配置中读取指定 key 到 target。
func UnmarshalKey(key string, target any) {
	if err := v.UnmarshalKey(key, target); err != nil {
		log.Printf("[WARN] unmarshal key %s failed: %v\n", key, err)
	}
}
