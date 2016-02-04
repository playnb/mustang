package sqldb

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/playnb/mustang/global"
)

func init() {
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	err := orm.RegisterDataBase("default", "mysql", global.C.MySqlUrl)
	if err != nil {
		fmt.Println(err)
	}
}

var Ormer orm.Ormer

func Init() {
	orm.RunSyncdb("default", false, true)

	Ormer = orm.NewOrm()
	Ormer.Using("default")
}
