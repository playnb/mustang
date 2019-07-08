package mgo_helper

import "gopkg.in/mgo.v2/bson"

//向数组里面加入数据，如果数组里已经存在，则不会加入（避免重复）
func AddToSet(m bson.M) bson.M {
	return bson.M{"$addToSet": m}
}

//删除数组元素，将所有匹配的元素删除
func Pull(m bson.M) bson.M {
	return bson.M{"$pull": m}
}

//向已有的数组末尾加入一个元素，要是元素不存在，就会创建一个新的元素。
func Push(m bson.M) bson.M {
	return bson.M{"$push": m}
}

//删除数组元素，只能从头部或尾部删除一个元素
func Pop(m bson.M) bson.M {
	return bson.M{"$pop": m}
}

func Set(m bson.M) bson.M {
	return bson.M{"$set": m}
}

func Inc(m bson.M) bson.M {
	return bson.M{"$inc": m}
}

func CheckObjectID(m string) bson.ObjectId {
	if len(m) == 24 {
		return bson.ObjectIdHex(m)
	}
	return bson.ObjectId("")
}
