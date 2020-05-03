package env

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/spf13/viper"
	"log"
	"os"
	"reflect"
	"strings"
)

func InitViper() {

	//viper.SetConfigFile("config.toml")
	viper.SetConfigName("config") //指定配置文件的文件名称(不需要制定配置文件的扩展名)
	viper.AddConfigPath(".")      // 设置配置文件和可执行二进制文件在用一个目录
	viper.AutomaticEnv()          //自动从环境变量读取匹配的参数

	//读取-c输入的路径参数，初始化配置文件，如： ./main -c config.yaml
	if len(os.Args) >= 3 {
		if os.Args[1] == "-c" {
			cfgFile := os.Args[2]
			viper.SetConfigFile(cfgFile)
		}
	}
	// 根据以上配置读取加载配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

//配置文件结构体，配置文件上的内容需要一一对应，可多不可少
type Configure struct {
	App struct {
		Name       string `json:"name" remark:"应用名称"`
		Version    string `json:"version" remark:"软件发布版本，对应仓库tag版本"`
		Mode       string `json:"mode" remark:"开发模式develop/test/product"`
		PRest      bool   `json:"prest" remark:"" remark:"是否开启pRest接口服务"`
		InitModels bool   `json:"init_models" remark:"是否初始化数据库模型" must:"false"`
	}
	Secret struct {
		JWT string `json:"jwt" remark:"jwt密钥"`
	}
	Rpc struct {
		Addr string `json:"addr" remark:"rpc主机地址"`
	}
	Db struct {
		Type     string `json:"type" remark:"数据库类型"`
		Host     string `json:"host" remark:"数据库主机"`
		Port     int64  `json:"port" remark:"数据库端口"`
		User     string `json:"user" remark:"数据库用户"`
		Password string `json:"password" remark:"数据库密码"`
		Dbname   string `json:"dbname" remark:"数据库名"`
		SslMode  string `json:"sslmode" remark:"ssl模式"`
	}
	Redis struct {
		Host     string `json:"host" remark:"redis主机"`
		Port     int64  `json:"port" remark:"redis端口"`
		Password string `json:"password" remark:"redis密码" must:"false"`
	}
}

var Conf = &Configure{}

//初始化配置信息，测试需要修改配置文件
func InitConfigure() (err error) {
	InitViper()

	confValue := reflect.ValueOf(Conf).Elem()
	confType := reflect.TypeOf(*Conf)

	for i := 0; i < confType.NumField(); i++ {
		section := confType.Field(i)
		sectionValue := confValue.Field(i)

		//读取节类型信息
		for j := 0; j < section.Type.NumField(); j++ {
			key := section.Type.Field(j)
			keyValue := sectionValue.Field(j)

			sec := strings.ToLower(section.Name) //配置文件节名
			remark := key.Tag.Get("remark")      //配置备注
			must := key.Tag.Get("must")          //配置备注
			tag := key.Tag.Get("json")           //配置键节名
			if tag == "" {
				err = fmt.Errorf("can not found a tag name `json` in struct of [%s].%s", sec, tag)
				logs.Error(err)
				return err
			}

			//绑定环境变量，会优先使用环境变量的值
			logs.Info("绑定环境变量 GZHUPI_%s_%s ==> %s.%s", strings.ToUpper(sec), strings.ToUpper(tag), sec, tag)
			envKey := fmt.Sprintf("GZHUPI_%s_%s", strings.ToUpper(sec), strings.ToUpper(tag))
			_ = viper.BindEnv(sec+"."+tag, envKey)

			//根据类型识别配置字段
			switch key.Type.Kind() {
			case reflect.String:
				value := viper.GetString(sec + "." + tag)
				if value == "" && must != "false" {
					err = fmt.Errorf("get a blank value of must item [%s].%s %s", sec, tag, remark)
					logs.Error(err)
					return err
				}
				keyValue.SetString(value)

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value := viper.GetInt64(sec + "." + tag)
				if value == 0 && must != "false" {
					err = fmt.Errorf("get a zero value of must item [%s].%s %s", sec, tag, remark)
					logs.Error(err)
					return err
				}
				keyValue.SetInt(value)

			case reflect.Bool:
				value := viper.GetBool(sec + "." + tag)
				keyValue.SetBool(value)

			default:
				logs.Warn("unsupported config struct key type %T", key.Type.Kind())
			}
		}
	}
	return
}
