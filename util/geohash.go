package util

import (
	"fmt"
	"strconv"
	//貌似API变了
	//"github.com/tidwall/tile38/client"
)

func Uint64ToString(number uint64) string {
	return fmt.Sprintf("%d", number)
}

func StringToUint64(str string) uint64 {
	f, _ := strconv.ParseUint(str, 10, 64)
	return f
}

func Float32ToString(number float32) string {
	return fmt.Sprintf("%f", number)
}

func StringToFloat32(str string) float32 {
	f, _ := strconv.ParseFloat(str, 32)
	return float32(f)
}

func Set(user_id uint64, x float32, y float32) (string, string, error) {
	/*
	pool, err := client.DialPool("localhost:9851")
	if err != nil {
		return "", "", nil
	}
	defer pool.CloseSession()

	var msg []byte
	var SetVar string
	conn, err := pool.Get()
	if err == nil {
		defer conn.CloseSession()

		SetVar = "set fleet " + Uint64ToString(user_id) + " point " + Float32ToString(y) + " " + Float32ToString(x)
		msg, err = conn.Do(SetVar)
	}

	return SetVar, string(msg), err
	*/

	return "", "", nil
}

func Get(user_id uint64) string {
	/*
	pool, err := client.DialPool("localhost:9851")
	if err != nil {
		return ""
	}
	defer pool.CloseSession()

	conn, err := pool.Get()
	if err != nil {
		return ""
	}
	defer conn.CloseSession()

	GetVar := "get fleet " + Uint64ToString(user_id) + " point"
	//resp, err := conn.Do("get fleet truck1 point")
	resp, _ := conn.Do(GetVar)
	return string(resp)
	*/

	return ""
}

func Nearby(x float32, y float32, distance uint64) (string, []uint64) {
	user_ids := make([]uint64, 0, 100)
	/*
			pool, err := client.DialPool("localhost:9851")
			if err != nil {
				return "", user_ids
			}
			defer pool.CloseSession()

			conn, err := pool.Get()
			if err != nil {
				return "", user_ids
			}
			defer conn.CloseSession()

			GetVar := "nearby fleet point " + Float32ToString(y) + " " + Float32ToString(x) + " " + Uint64ToString(distance)
			//resp, err := conn.Do("get fleet truck1 point")
			resp, _ := conn.Do(GetVar)

			ss := string(resp)

			ss1 := strings.Split(ss, ",")
			for _, ss2 := range ss1 {
				ss3 := strings.Split(ss2, "{")
				if len(ss3) != 2 {
					continue
				}

				for _, ss4 := range ss3 {
					ss5 := strings.Split(ss4, ":")
					if len(ss5) != 2 {
						continue
					}

					key_str := strings.Replace(ss5[0], "\"", "", -1)
					if strings.Compare(key_str, "id") != 0 {
						continue
					}

					value_str := strings.Replace(ss5[1], "\"", "", -1)
					user_ids = append(user_ids, StringToUint64(value_str))
				}
			}

			return string(resp), user_ids
		*/
	return "", user_ids
}
