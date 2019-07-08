package tbl

import (
	"github.com/playnb/mustang/log"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

type interfaceXlsxData interface {
	GetUniqueID() uint64 //计算唯一ID
	AfterLoad()          //加载完成
	OnLoad(name string)  //加载字段name(struct的变量名)
}

//LoadXlsx 加载xlsx文件
//TODO: 对于[]uint64 map[uint64]uint64等的容器类型,需要在顶一个tag表示分隔符
func LoadXlsx(excelFileName string, sheetName string, dataType interfaceXlsxData) interface{} {
	dm := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(uint64(0)), reflect.TypeOf(dataType)))

	xlFile, err := xlsx.OpenFile(excelFileName)

	count := 0
	for count < 6 && err != nil {
		log.Trace("重新读取配置文件 %s 上次加载失败 %s", excelFileName, err)
		time.Sleep(time.Second * 10)
		xlFile, err = xlsx.OpenFile(excelFileName)
		count++
	}

	if err != nil {
		log.Trace("加载配置文件 %s 失败 %s", excelFileName, err)
		return nil
	}

	log.Trace("加载配置文件 %s ==> %s", excelFileName, sheetName)
	foundSheet := false
	for _, sheet := range xlFile.Sheets {
		if sheet.Name == sheetName {
			foundSheet = true
			colName := make(map[int]string)
			tableName := make(map[string]bool)
			{
				s := reflect.New(reflect.TypeOf(dataType).Elem()).Interface().(interfaceXlsxData)
				e := reflect.TypeOf(s).Elem()
				for i := 0; i < e.NumField(); i++ {
					tn := e.Field(i).Tag.Get("xlsx")
					if len(tn) > 0 {
						tableName[tn] = true
					}
				}
			}
			for rowIndex, row := range sheet.Rows {
				if rowIndex == 0 {
					for colIndex, cell := range row.Cells {
						if tableName[cell.Value] {
							colName[colIndex] = cell.Value
						}
					}

					for k := range tableName {
						found := false
						for _, v := range colName {
							if k == v {
								found = true
								break
							}
						}
						if found == false {
							log.Error("===== 读取配置 %s,%s 表格缺少字段[%s]", excelFileName, sheetName, k)
						}
					}
				} else if rowIndex == 1 {
				} else if rowIndex == 2 {
				} else {
					//log.Debug("读取行 %s", row.Cells[0])
					cells := make(map[string]*(xlsx.Cell))
					for colIndex, cell := range row.Cells {
						if _, ok := colName[colIndex]; ok {
							cells[colName[colIndex]] = cell
						}
						//log.Debug("%s: %s", colName[colIndex], cell.Value)
					}

					dataReady := false
					data := reflect.New(reflect.TypeOf(dataType).Elem()).Interface().(interfaceXlsxData)
					s := reflect.TypeOf(data).Elem()
					for i := 0; i < s.NumField(); i++ {
						tagName := s.Field(i).Tag.Get("xlsx")
						sep1 := s.Field(i).Tag.Get("sep1")
						sep2 := s.Field(i).Tag.Get("sep2")
						if len(tagName) == 0 {
							continue
						}
						//log.Debug("%d: %s", i, tagName)
						if s.Field(i).Type.Kind() == reflect.Ptr {
							v := reflect.ValueOf(data).Elem()
							v.Field(i).Set(reflect.New(s.Field(i).Type.Elem()))
						}
						if cell, ok := cells[tagName]; ok {
							if len(cell.Value) > 0 {
								dataReady = true
							}
							switch s.Field(i).Type.Kind() {
							case reflect.Int:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetInt(0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetInt(value)
								}
							case reflect.Int32:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetInt(0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetInt(value)
								}
							case reflect.Uint32:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetUint(0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetUint(uint64(value))
								}
							case reflect.Int64:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetInt(0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetInt(value)
								}
							case reflect.Uint64:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetUint(0)
									//fmt.Println(err)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetUint(uint64(value))
								}
							case reflect.String:
								value, err := cell.String()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetString("")
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetString(value)
								}
							case reflect.Bool:
								value, err := cell.Int64()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetBool(false)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetBool(value != 0)
								}
							case reflect.Float32:
								value, err := cell.Float()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetFloat(0.0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetFloat(value)
								}
							case reflect.Float64:
								value, err := cell.Float()
								if err != nil {
									reflect.ValueOf(data).Elem().Field(i).SetFloat(0.0)
								} else {
									reflect.ValueOf(data).Elem().Field(i).SetFloat(value)
								}
							case reflect.Map:
								if len(sep1) == 0 {
									sep1 = ";"
								}
								if len(sep2) == 0 {
									sep2 = ":"
								}
								typeName := s.Field(i).Type.String()
								if typeName == "map[uint64]uint64" {
									if reflect.ValueOf(data).Elem().Field(i).IsNil() {
										reflect.ValueOf(data).Elem().Field(i).Set(
											reflect.MakeMap(reflect.MapOf(reflect.TypeOf(uint64(0)), reflect.TypeOf(uint64(0)))))
									}
									value, _ := cell.String()
									ss1 := strings.Split(value, sep1)
									for _, s1 := range ss1 {
										ss2 := strings.Split(s1, sep2)
										if len(ss2) == 2 {
											nk, err1 := strconv.ParseUint(ss2[0], 10, 64)
											nv, err2 := strconv.ParseUint(ss2[1], 10, 64)
											if err1 == nil && err2 == nil {
												reflect.ValueOf(data).Elem().Field(i).SetMapIndex(reflect.ValueOf(nk), reflect.ValueOf(nv))
											}
										}
									}
								} else if typeName == "map[uint64]struct {}" {
									if reflect.ValueOf(data).Elem().Field(i).IsNil() {
										reflect.ValueOf(data).Elem().Field(i).Set(
											reflect.MakeMap(reflect.MapOf(reflect.TypeOf(uint64(0)), reflect.TypeOf(struct{}{}))))
									}
									value, _ := cell.String()
									ss1 := strings.Split(value, sep1)
									for _, s1 := range ss1 {
										nk, err := strconv.ParseUint(s1, 10, 64)
										if err == nil {
											reflect.ValueOf(data).Elem().Field(i).SetMapIndex(reflect.ValueOf(nk), reflect.ValueOf(struct{}{}))
										}
									}
									//log.Debug("======================>>>>>>>>> ===> %v", reflect.ValueOf(data).Elem().Field(i))
								}
								//log.Debug("======> %s", typeName)
							case reflect.Slice:
								if len(sep1) == 0 {
									sep1 = ";"
								}
								typeName := s.Field(i).Type.String()
								if typeName == "[]uint64" {
									if reflect.ValueOf(data).Elem().Field(i).IsNil() {
										reflect.ValueOf(data).Elem().Field(i).Set(
											reflect.MakeSlice(reflect.TypeOf(make([]uint64, 0, 0)), 0, 10))
									}
									value, _ := cell.String()
									ss1 := strings.Split(value, sep1)
									for _, s1 := range ss1 {
										nk, err := strconv.ParseUint(s1, 10, 64)
										if err == nil {
											reflect.ValueOf(data).Elem().Field(i).Set(
												reflect.Append(
													reflect.ValueOf(data).Elem().Field(i),
													reflect.ValueOf(nk),
												),
											)
										}
									}
									//log.Debug("======================>>>>>>>>> ===> %v", reflect.ValueOf(data).Elem().Field(i))
								} else if typeName == "[]float32" {
									if reflect.ValueOf(data).Elem().Field(i).IsNil() {
										reflect.ValueOf(data).Elem().Field(i).Set(
											reflect.MakeSlice(reflect.TypeOf(make([]float32, 0, 0)), 0, 10))
									}
									value, _ := cell.String()
									ss1 := strings.Split(value, sep1)
									for _, s1 := range ss1 {
										nk, err := strconv.ParseFloat(s1, 64)
										if err == nil {
											reflect.ValueOf(data).Elem().Field(i).Set(
												reflect.Append(
													reflect.ValueOf(data).Elem().Field(i),
													reflect.ValueOf(float32(nk)),
												),
											)
										}
									}
									//log.Debug("======================>>>>>>>>> ===> %v", reflect.ValueOf(data).Elem().Field(i))
								} else if typeName == "[]string" {
									if reflect.ValueOf(data).Elem().Field(i).IsNil() {
										reflect.ValueOf(data).Elem().Field(i).Set(
											reflect.MakeSlice(reflect.TypeOf(make([]string, 0, 0)), 0, 10))
									}
									value, _ := cell.String()
									ss1 := strings.Split(value, sep1)
									for _, s1 := range ss1 {
										strings.Trim(s1, " ")
										if len(s1) > 0 {
											reflect.ValueOf(data).Elem().Field(i).Set(
												reflect.Append(
													reflect.ValueOf(data).Elem().Field(i),
													reflect.ValueOf(s1),
												),
											)
										}
									}
								}
							case reflect.Ptr:
								v := reflect.ValueOf(data).Elem()
								f := v.Field(i).MethodByName("ParseStr")
								if f.IsValid() && f.Kind() == reflect.Func && f.Type().NumIn() == 1 {
									val, _ := cell.String()
									f.Call([]reflect.Value{reflect.ValueOf(val)})
								} else {
									log.Debug("====================== unknown type %d ===> %s", s.Field(i).Type.Kind(), s.Field(i).Type.String())
								}
							default:
								typeName := s.Field(i).Type.String()
								if typeName == "time.Time" {
									if sep1 == "" {
										sep1 = "2006-1-2 15:04:05"
									}
									value, err := cell.String()

									var timeValue time.Time
									if err != nil {
										timeValue = time.Unix(0, 0)
									} else {
										//timeValue, err = time.Parse(sep1, value)
										timeValue, err = time.ParseInLocation(sep1, value, time.Local)
										if err != nil {
											timeValue = time.Unix(0, 0)
											log.Error("加载Excel失败 类型time %s: %s", value, err)
										}
									}
									reflect.ValueOf(data).Elem().Field(i).Set(reflect.ValueOf(timeValue))
								} else {
									log.Debug("====================== unknown type %d ===> %s", s.Field(i).Type.Kind(), typeName)
								}
							}
							data.OnLoad(s.Field(i).Name)

							//log.Debug("SSS> %d: %s  ==> %s  | %v", i, tagName, cell.Value, data)
						} else {
							str := ""
							for k, v := range cells {
								str = str + fmt.Sprintf("%v,%v| ", k, v)
							}
							//TODO: 这里处理默认数值
							//log.Warning("读取配置 %s,%s 无效字段[%s] ==> %s", excelFileName, sheetName, tagName, str)

							//dataReady = false
						}
					}
					if dataReady == true {
						data.AfterLoad()
						dm.SetMapIndex(reflect.ValueOf(data.GetUniqueID()), reflect.ValueOf(data))
					} else {
						log.Warning("跳过数据 %d: %v ", rowIndex, data)
					}
				}
			}
		}
	}
	if foundSheet == false {
		log.Error("===== 读取配置 %s,%s 表格没有相应的sheet", excelFileName, sheetName)
	}
	return dm.Interface()
}
