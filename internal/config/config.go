package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 只保留启动前必须确定的参数。
// 其余配置（实例列表、数据保留天数、GeoIP 路径）在运行时从数据库读写。
type Config struct {
	Listen  string `yaml:"listen"`
	DataDir string `yaml:"data_dir"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Listen:  ":8080",
		DataDir: "./data",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // 配置文件不存在时使用默认值
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
