package util_db

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/util"
	"github.com/go-redis/redis"
	"reflect"
	"strconv"
)

/*
针对这种结构的数据 tag:redis
type Zt2User struct {
	ID          uint64 `redis:"charid"`
	Name        string `redis:"name"`
}
写入Redis的Hash中
database: 目标数据库
key: 主键
data: 数据
dbCharSet: Redis数据库中存放的字符编码
*/
func RedisHSet(database *redis.Client, key string, data interface{}, dbCharSet util.Charset) {
	query := make(map[string]interface{})
	for i := 0; i < reflect.TypeOf(data).Elem().NumField(); i++ {
		t := reflect.TypeOf(data).Elem().Field(i)
		v := reflect.ValueOf(data).Elem().Field(i)
		m := t.Tag.Get("redis")
		if len(m) > 0 {
			if t.Type.Kind() == reflect.String {
				if util.GBK == dbCharSet {
					query[m] = util.ConvertUTF8Bytes(v.String())
				} else {
					query[m] = reflect.ValueOf(data).Elem().Field(i).Interface()
				}
			} else if t.Type.Kind() == reflect.Ptr {
				if v.IsNil() {
					continue
				}
				if f := v.MethodByName("Marshal"); f.IsValid() && f.Type().NumIn() == 0 && f.Type().NumOut() == 2 {
					ret := f.Call(nil)
					if len(ret) != 2 || ret[0].IsNil() {
						log.Error("RedisHGet: Marshal错误")
					} else {
						query[m] = ret[0].Interface()
					}
				}
			} else {
				query[m] = reflect.ValueOf(data).Elem().Field(i).Interface()
			}
		}
	}
	database.HMSet(key, query)
	//fmt.Println(database.HMSet(key, query))
}

//TODO 这里不需要每次都reflect一次，学学人家Protobuf的处理，做个map缓存
func makeHMember(dst interface{}) ([]string, []reflect.Value) {
	V := reflect.ValueOf(dst)
	if V.Type().Kind() == reflect.Ptr {
		V = V.Elem()
	}
	var query []string
	var fields []reflect.Value
	for i := 0; i < V.NumField(); i++ {
		t := V.Type().Field(i)
		if t.Type.Kind() == reflect.Struct && t.Anonymous {
			q, f := makeHMember(V.Field(i).Addr().Interface())
			query = append(query, q...)
			fields = append(fields, f...)
		} else {
			m := t.Tag.Get("redis")
			if len(m) > 0 {
				query = append(query, m)
				fields = append(fields, V.Field(i))
			}
		}
	}

	return query, fields
}

type redisHMGet interface {
	HMGet(key string, fields ...string) *redis.SliceCmd
}

func RedisHGet(database redisHMGet, key string, dst interface{}, dbCharSet util.Charset) bool {
	/*
		var query []string
		var fields []reflect.Value
		for i := 0; i < reflect.TypeOf(dst).Elem().NumField(); i++ {
			t := reflect.TypeOf(dst).Elem().Field(i)
			if t.Type.Kind() == reflect.Struct && t.Anonymous {

			}
			m := t.Tag.Get("redis")
			if len(m) > 0 {
				query = append(query, m)
				fields = append(fields, reflect.ValueOf(dst).Elem().Field(i))
			}
		}
	*/
	query, fields := makeHMember(dst)
	data := database.HMGet(key, query...).Val()
	found := false
	if len(data) != len(fields) {
		return false
	}
	for k, v := range fields {
		//fmt.Println(t.Name + " " + t.Tag.Get("redis"))
		value := ""
		if data[k] == nil {
			continue
		}
		found = true

		if v.Kind() == reflect.String {
			if reflect.TypeOf(data[k]).Kind() == reflect.String {
				if util.GBK == dbCharSet {
					value = util.ConvertGBKString([]byte(data[k].(string)))
				} else {
					value = data[k].(string)
				}
			} else {
				continue
			}
		} else {
			if reflect.TypeOf(data[k]).Kind() == reflect.String {
				value = data[k].(string)
			} else {
				continue
			}
		}

		switch v.Kind() {
		case reflect.Int64:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int:
			n, _ := strconv.ParseInt(value, 10, 64)
			v.SetInt(n)
		case reflect.Uint64:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint:
			n, _ := strconv.ParseUint(value, 10, 64)
			v.SetUint(n)
		case reflect.Float64:
			fallthrough
		case reflect.Float32:
			n, _ := strconv.ParseFloat(value, 64)
			v.SetFloat(n)
		case reflect.String:
			v.SetString(value)
		case reflect.Bool:
			n, _ := strconv.ParseBool(value)
			v.SetBool(n)
		case reflect.Slice:
			//if v.Type().String() == "[]bytes" {
			//}
		case reflect.Struct:
		case reflect.Ptr:
			if f := v.MethodByName("Unmarshal"); f.IsValid() && f.Type().NumIn() == 1 && f.Type().NumOut() == 1 {
				v.Set(reflect.New(v.Type().Elem()))
				err := f.Call([]reflect.Value{reflect.ValueOf([]byte(value))})[0]
				if err.IsNil() == false {
					log.Error("RedisHGet: %s", err.String())
				}
			} else {
				log.Error("RedisHGet: 遇到不支持的结构")
			}
		case reflect.Interface:
		}
	}
	return found
}
