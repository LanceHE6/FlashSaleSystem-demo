package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

// 定义数据结构

type Config struct {
	SERVER struct {
		PORT string `yaml:"port"`
		MODE string `yaml:"mode"`
	} `yaml:"server"`

	MYSQL struct {
		HOST     string `yaml:"host"`
		PORT     string `yaml:"port"`
		ACCOUNT  string `yaml:"account"`
		PASSWORD string `yaml:"password"`
		DBNAME   string `yaml:"dbname"`
	} `yaml:"mysql"`

	REDIS struct {
		HOST     string `yaml:"host"`
		PORT     string `yaml:"port"`
		PASSWORD string `yaml:"password"`
		DBNAME   int    `yaml:"dbname"`
	} `yaml:"redis"`

	RABBITMQ struct {
		HOST     string `yaml:"host"`
		PORT     string `yaml:"port"`
		ACCOUNT  string `yaml:"account"`
		PASSWORD string `yaml:"password"`
		VHOST    string `yaml:"vhost"`
	} `yaml:"rabbitmq"`
}

var ServerConfig Config

// 创建一个函数来读取和解析YAML文件

func init() {
	// 检查文件是否存在
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		// 如果文件不存在，创建并写入默认值
		defaultConfig := Config{
			SERVER: struct {
				PORT string `yaml:"port"`
				MODE string `yaml:"mode"`
			}{
				PORT: "8080",
				MODE: "debug",
			},

			MYSQL: struct {
				HOST     string `yaml:"host"`
				PORT     string `yaml:"port"`
				ACCOUNT  string `yaml:"account"`
				PASSWORD string `yaml:"password"`
				DBNAME   string `yaml:"dbname"`
			}{
				HOST:     "localhost",
				PORT:     "3306",
				ACCOUNT:  "root",
				PASSWORD: "root",
				DBNAME:   "flash_sale",
			},

			REDIS: struct {
				HOST     string `yaml:"host"`
				PORT     string `yaml:"port"`
				PASSWORD string `yaml:"password"`
				DBNAME   int    `yaml:"dbname"`
			}{
				HOST:     "localhost",
				PORT:     "6379",
				PASSWORD: "123456",
				DBNAME:   0,
			},

			RABBITMQ: struct {
				HOST     string `yaml:"host"`
				PORT     string `yaml:"port"`
				ACCOUNT  string `yaml:"account"`
				PASSWORD string `yaml:"password"`
				VHOST    string `yaml:"vhost"`
			}{
				HOST:     "localhost",
				PORT:     "5672",
				ACCOUNT:  "hycer",
				PASSWORD: "123456",
				VHOST:    "my_vhost",
			},
		}

		defaultBytes, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			_ = fmt.Errorf("can not marshal the default config")
			return
		}

		err = ioutil.WriteFile("config.yaml", defaultBytes, 0644)
		if err != nil {
			_ = fmt.Errorf("can not write the default config to file")
			return
		}
	}

	// 读取文件
	fileBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		_ = fmt.Errorf("can not read the config file")
		return
	}

	// 解析YAML文件
	err = yaml.Unmarshal(fileBytes, &ServerConfig)
	if err != nil {
		_ = fmt.Errorf("can not unmarshal the config file")
		return
	}

}
