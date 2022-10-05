package main

func _builtin_add(self *TsapaObject, machine *TsapaMachine) *TsapaObject {
	if machine.tempScope["__x"].type_str == NumericType {
		return constructObject(NumericType, machine.tempScope["__x"].value().(int)+machine.tempScope["__y"].value().(int))
	} else if machine.tempScope["__x"].type_str == StringType {
		return constructObject(StringType, machine.tempScope["__x"].value().(string)+machine.tempScope["__y"].value().(string))
	} else if machine.tempScope["__x"].type_str == FloatType {
		return constructObject(FloatType, machine.tempScope["__x"].value().(float64)+machine.tempScope["__y"].value().(float64))
	}
	panic("native function add: -- type mismatch")
	return nullObject()
}

func _builtin_sub(self *TsapaObject, machine *TsapaMachine) *TsapaObject {
	if machine.tempScope["__x"].type_str == NumericType {
		return constructObject(NumericType, machine.tempScope["__x"].value().(int)-machine.tempScope["__y"].value().(int))
	} else if machine.tempScope["__x"].type_str == FloatType {
		return constructObject(FloatType, machine.tempScope["__x"].value().(float64)-machine.tempScope["__y"].value().(float64))
	}
	panic("native function sub: -- type mismatch")
	return nullObject()
}

func _builtin_inc(self *TsapaObject, machine *TsapaMachine) *TsapaObject {
	if machine.tempScope["__x"].type_str == NumericType {
		return constructObject(NumericType, machine.tempScope["__x"].value().(int)+1)
	}
	panic("native function inc: -- type mismatch")
	return nullObject()
}

func constructBuiltinFunction(args []string, handler func(self *TsapaObject, machine *TsapaMachine) *TsapaObject) *TsapaObject {
	obj := constructObject(CallableType, nil)
	obj.fields["_args"] = constructObject(_args_arr, args)
	obj._native_eval_hook = handler
	return obj
}

func constructBuiltinsScope() map[string]*TsapaObject {
	dict := make(map[string]*TsapaObject)
	dict["none"] = nullObject()

	dict["add"] = constructBuiltinFunction([]string{"__x", "__y"}, _builtin_add)
	dict["sub"] = constructBuiltinFunction([]string{"__x", "__y"}, _builtin_sub)
	dict["inc"] = constructBuiltinFunction([]string{"__x"}, _builtin_inc)

	return dict
}
