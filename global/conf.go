package global

import (
	"flag"
	"fmt"
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/utils"
	"strconv"
)

//取得钻石配置
type AchievementType struct {
	Score    string `json:"score"`
	Diamonds string `json:"diamonds"`
}

func (a *AchievementType) GetScore() uint32 {
	v, _ := strconv.Atoi(a.Score)
	return uint32(v)
}
func (a *AchievementType) GetDiamonds() uint32 {
	v, _ := strconv.Atoi(a.Diamonds)
	return uint32(v)
}

//商店商品配置
type ShopItemType struct {
	Character     string `json:"character"`
	Diamonds string `json:"diamonds"`
	Level    string `json:"level"`
}

func (s *ShopItemType) GetLevel() uint32 {
	v, _ := strconv.Atoi(s.Level)
	return uint32(v)
}
func (s *ShopItemType) GetDiamonds() uint32 {
	v, _ := strconv.Atoi(s.Diamonds)
	return uint32(v)
}

type Config struct {
	AppID         string `json:"AppID"`
	AppSecret     string `json:"AppSecret"`
	OwnerID       string `json:"OwnerID"`
	ServiceDomain string `json:"ServiceDomain"`

	RedisUrl string `json:"RedisUrl"`
	RedisPin string `json:"RedisPin"`

	MySqlUrl string `json:"MySqlUrl"`

	SecretKey string `json:"SecretKey"`
	Nonce     string `json:"Nonce"`

	HttpPort string `json:"HttpPort"`
	AppPort  string `json:"AppPort"`

	ServiceToken string `json:"ServiceToken"`

	LoginTimeOut int `json:"LoginTimeOut"`

	Release int `json:"Release"`

	DiamondsByFirend uint32            `json:"DiamondsByFirend"`
	Achievement      []AchievementType `json:"Achievement"`
	ShopItem         []ShopItemType    `json:"ShopItem"`
}

var C = &Config{}
var Release = false

func LoadConfig() {
	config_file := flag.String("config", "", "Use -config <filesource>")
	config_url := flag.String("config_url", "", "Use -config_url <filesource>")

	flag.Parse()

	utils.GetParentDirectory(utils.GetCurrentDirectory())

	if len(*config_file) > 1 {
		fmt.Println("读取配置文件: " + *config_file)
		utils.LoadJsonFile(*config_file, C)
	} else if len(*config_url) > 1 {
		fmt.Println("读取配置文件: " + *config_url)
		utils.LoadJsonURL(*config_url, C)
	}
	log.Release = (C.Release != 0)
	Release = (C.Release != 0)
	if Release {
		log.Trace("Release模式运行")
	} else {
		log.Trace("Debug模式运行")
	}
	//一些特殊值
	if C.LoginTimeOut <= 1 {
		C.LoginTimeOut = 3
	}
	log.Debug("%v", C)
}

func init() {
	LoadConfig()
}
