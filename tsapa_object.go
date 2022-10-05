package main

import (
	"fmt"
	"strconv"
)

type TsapaObject struct {
	type_str          int
	fields            map[string]*TsapaObject
	_value            interface{}
	_native_eval_hook func(self *TsapaObject, machine *TsapaMachine) *TsapaObject
}

const ( //type_str is int for reasons of performance and memory saving
	NoneType    = iota
	NumericType //matches to "int"
	FloatType   //"float64"
	StringType  //"string"
	StructType
	CallableType //"string" (when reflecting non-native functions)
	BooleanType  //"bool"
	_args_arr
	// in CallableType objects, _value must be an array of strings containing sentences,
	// and "_args" field must be an array of strings containing names of method arguments
)

func (obj *TsapaObject) toString() string {
	switch obj.type_str {
	case NoneType:
		return "None"
	case NumericType:
		return strconv.Itoa(obj.value().(int))
	case FloatType:
		return fmt.Sprintf("%f", obj.value().(float64))
	case StringType:
		return obj.value().(string)
	case StructType:
		return "<StructType: object with fields>"
	case CallableType:
		return "<CallableType>"
	case BooleanType:
		if obj.value().(bool) {
			return "true"
		} else {
			return "false"
		}
	case _args_arr:
		return fmt.Sprintf("%v", obj.value().([]string))
	}
	return "Unknown type"
}

func (obj *TsapaObject) value() interface{} {
	return obj._value
}

func (obj *TsapaObject) eval(machine *TsapaMachine, rec int) *TsapaObject {
	if obj._native_eval_hook == nil {
		return machine._eval(obj._value.(string), rec+1) //assuming that _value in CallableType is string
	} else {
		return obj._native_eval_hook(obj, machine)
	}
}

func constructObject(type_str int, value interface{}) *TsapaObject {
	obj := new(TsapaObject)
	obj.type_str = type_str
	obj.fields = make(map[string]*TsapaObject)
	obj._native_eval_hook = nil
	if type_str != NoneType {
		obj._value = value
	}
	return obj
}

func nullObject() *TsapaObject {
	return constructObject(NoneType, nil)
}

func typeStringMap() map[int]string {
	return map[int]string{
		NoneType:     "NoneType",
		NumericType:  "NumericType",
		FloatType:    "FloatType",
		StringType:   "StringType",
		StructType:   "StructType",
		CallableType: "CallableType",
		BooleanType:  "BooleanType",
		_args_arr:    "_args_arrType",
	}
}
