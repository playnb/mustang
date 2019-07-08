package util

import "reflect"

/*
DelSlice 删除slice中满足filter条件的元素
slice 传入切片的引用,否则会panic!!!
filter 删除条件
*/
func DelSlice(slice interface{}, filter func(d interface{}) bool) {
	var rv reflect.Value
	rv = reflect.ValueOf(slice).Elem()

	length := rv.Len()
	for i := 0; i < length; {
		if filter(rv.Index(i).Interface()) {
			rv = reflect.AppendSlice(rv.Slice(0, i), rv.Slice(i+1, length))
			length = rv.Len()
		} else {
			i++
		}
	}
	reflect.ValueOf(slice).Elem().Set(rv)
}

//slice 传入切片的引用,否则会panic!!!
func DelSliceIndex(slice interface{}, index int) {
	rv := reflect.ValueOf(slice).Elem()
	length := rv.Len()
	if index >= rv.Len() {
		return
	}
	rv = reflect.AppendSlice(rv.Slice(0, index), rv.Slice(index+1, length))
	reflect.ValueOf(slice).Elem().Set(rv)
}

func SubSlice(slice interface{}, begin int, end int) interface{} {
	if begin > end {
		t := begin
		begin = end
		end = t
	}
	rv := reflect.ValueOf(slice)
	length := rv.Len()
	if begin > length {
		begin = length
	}
	if end > length {
		end = length
	}
	return rv.Slice(begin, end).Interface()
}

func FindSlice(slice interface{}, filter func(d interface{}) bool) interface{} {
	var rv reflect.Value
	rv = reflect.ValueOf(slice)

	length := rv.Len()
	for i := 0; i < length; i++ {
		if filter(rv.Index(i).Interface()) {
			return rv.Index(i).Interface()
		}
	}
	return nil
}
