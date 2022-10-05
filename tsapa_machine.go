package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type TsapaRegexpSet struct {
	// Set of Regexp objects for fast tokenizing.
	// Each regexp definition is followed by an example of structure it should\
	// be able to match.
	block      *regexp.Regexp // [x|x+1]
	comment    *regexp.Regexp // //foobar
	key        *regexp.Regexp // foo
	numeric    *regexp.Regexp // 42
	float      *regexp.Regexp // 42.13
	str        *regexp.Regexp // "foo"
	call       *regexp.Regexp // Console println:
	call_func  *regexp.Regexp // foo:bar
	parenth    *regexp.Regexp // (not:false)
	boolvar    *regexp.Regexp // false
	macro      *regexp.Regexp // !inspect
	assignment *regexp.Regexp // foo <- "bar"
	_extends   *regexp.Regexp // foo extends bar
	_args      *regexp.Regexp // [x,y|
	copy_obj   *regexp.Regexp // foo copy bar
	sep        *regexp.Regexp // foo; bar
	_str_val   *regexp.Regexp // "foo\"bar\""
	// unary      *regexp.Regexp // 2 + 2
}

func constructRegexpSet() *TsapaRegexpSet {
	obj := new(TsapaRegexpSet)
	obj.comment = regexp.MustCompile("^//.*")
	obj.key = regexp.MustCompile("^(\\w+)$")
	obj.block = regexp.MustCompile("^\\[([a-zA-Z_, ])*\\|(.*)\\]$")
	obj.numeric = regexp.MustCompile("^(-)?(\\d)+$")
	obj.float = regexp.MustCompile("^(-)?(\\d)+\\.(\\d)+$")
	obj.call = regexp.MustCompile("^([0-9A-Za-z_,\\s])+ ([a-zA-Z_, ]+:\\w*)+")
	obj.call_func = regexp.MustCompile("^[0-9a-zA-Z_,]+:\\w*")
	obj.parenth = regexp.MustCompile("^\\(.*\\)$")
	obj.boolvar = regexp.MustCompile("^(true|false)$")
	obj.macro = regexp.MustCompile("^!.+")
	obj.assignment = regexp.MustCompile("^(\\w)+ <- .+")
	obj._extends = regexp.MustCompile(" extends \\w+$")
	//FIXME []* with no args provided makes Tsapa runtime create NoneType object named " " in tempScope on function call
	obj._args = regexp.MustCompile("^\\[[0-9A-Za-z_,\\s]*\\|")
	obj.copy_obj = regexp.MustCompile("copy \\w+$")
	//obj.sep = regexp.MustCompile("^(([^\"]*;.*)|((.*\"){2};.*))+$")
	obj.sep = regexp.MustCompile("^.*;.*$")
	obj._str_val = regexp.MustCompile(`(([^\\]{1}|^)".*[^\\]{1}")|(([^\\]{1}|^)"")`)
	// obj.unary = regexp.MustCompile("^(\\w+|\\(.*\\))(\\s)*([+-/*//%]|<<|>>)(\\s)*(\\w+|\\(.*\\))$")
	// Original, unescaped version :
	// ^(\w+|\(.*\))(\s)*([+-/*//%]|<<|>>)(\s)*(\w+|\(.*\))$
	return obj
}

type TsapaMachine struct {
	variableScope map[string]*TsapaObject
	builtinScope  map[string]*TsapaObject
	tempScope     map[string]*TsapaObject
	re            *TsapaRegexpSet
	trace         bool
}

func constructMachine() *TsapaMachine {
	obj := new(TsapaMachine)
	obj.variableScope = make(map[string]*TsapaObject)
	obj.tempScope = make(map[string]*TsapaObject)
	obj.builtinScope = constructBuiltinsScope()
	obj.re = constructRegexpSet()
	obj.trace = false
	return obj
}

func (machine *TsapaMachine) process(statement string) string {
	defer func(stmt string) {
		if r := recover(); r != nil {
			fmt.Printf("Achtung! Runtime error: %v\n\t-> \"%s\"\n", r, stmt)
		}
	}(statement)

	t := machine._eval(statement, 0)

	return t.toString()
}

func (machine *TsapaMachine) _getObject(name string) *TsapaObject {
	if object, test := machine.tempScope[name]; test { //typical chinese code
		return object
	} else if object, test := machine.variableScope[name]; test {
		return object
	} else if object, test := machine.builtinScope[name]; test {
		return object
	} else {
		panic("no such object in current scope")
	}
}

func (machine *TsapaMachine) _eval(token string, rec int) *TsapaObject {
	if machine.trace {
		fmt.Printf("Depth: %d, token: \"%s\"\n", rec, token)
	}

	if strings.Contains(token, "\"") {
		str_objects := machine.re._str_val.FindAllString(token, -1)
		if str_objects == nil {
			panic("string cannot be parsed")
		} else {
			for _, str_obj := range str_objects {
				//internal_object := machine._eval(str_obj, rec+1)
				str_obj := strings.TrimSpace(str_obj)
				internal_object := constructObject(StringType, strings.Replace(strings.TrimSuffix(strings.TrimPrefix(str_obj, `"`), `"`), `\"`, `"`, -1))
				obj_name := "_s_" + fmt.Sprintf("%p", &internal_object)
				machine.tempScope[obj_name] = internal_object
				token = strings.Replace(token, str_obj, obj_name, 1)
			}
		}
	}

	if machine.re.comment.MatchString(token) {

		return nullObject()

	} else if machine.re.sep.MatchString(token) {

		var intermResult *TsapaObject
		statements := strings.Split(token, ";")

		for _, statement := range statements {
			intermResult = machine._eval(strings.TrimSpace(statement), rec+1)
		}

		return intermResult //TODO in theory, re.sep case should return an array of TsapaObjects, returned by corresponding statements in token

	} else if machine.re.assignment.MatchString(token) {

		if machine.re._extends.MatchString(token) {
			objname := strings.TrimLeft(regexp.MustCompile(" \\w+$").FindString(token), " ")

			object := machine._getObject(objname)

			cltok := strings.Split(machine.re._extends.ReplaceAllString(token, ""), " <- ")
			object.fields[cltok[0]] = machine._eval(cltok[1], rec+1)
			return constructObject(StringType, object.fields[cltok[0]].toString())
		} else {
			tokens := strings.Split(token, " <- ")
			machine.variableScope[tokens[0]] = machine._eval(tokens[1], rec+1)

			return constructObject(StringType, machine.variableScope[tokens[0]].toString())
		}

		return nullObject()

	} else if machine.re.call.MatchString(token) {

		subtokens := strings.SplitN(token, " ", 2) // SplitN() will return slice with two strings
		objectName := subtokens[0]
		callstrings := subtokens[1]

		object := machine._getObject(objectName)

		//----TEMPORARY----
		callstring := strings.SplitN(callstrings, ":", 2)
		fieldName := callstring[0]
		//args := callstring[1]
		if field, test := object.fields[fieldName]; test {
			if field.type_str == CallableType {
				//arguments should be passed along here
				return field.eval(machine, rec)
			} else {
				return field
			}
		} else {
			panic("no such field or method: " + objectName + ":" + fieldName)
		}

	} else if machine.re.call_func.MatchString(token) {

		callstring := strings.SplitN(token, ":", 2)
		funcName := callstring[0]
		args := callstring[1]

		func_obj := machine._getObject(funcName)

		if func_obj.type_str == CallableType {
			defer func() {
				if r := recover(); r != nil {
					panic(fmt.Sprintf("(arguments mismatch) %v", r))
				}
			}()

			args_tokens := strings.Split(args, ",")

			for i, token := range args_tokens {
				args_tokens[i] = strings.Trim(token, " ")
			}

			arg_names := func_obj.fields["_args"].value().([]string)

			if len(args_tokens) > len(arg_names) {
				panic("too many arguments") //FIXME : overridden by panic("arguments mismatch")
			}

			for index, argument_name := range arg_names {
				machine.tempScope[argument_name] = machine._eval(args_tokens[index], rec+1)
			}
			result := func_obj.eval(machine, rec)
			if rec == 0 {
				machine.tempScope = make(map[string]*TsapaObject)
			}
			return result
		} else {
			panic(funcName + " is not callable")
		}

	} else if machine.re.copy_obj.MatchString(token) {

		object := machine._getObject(strings.TrimPrefix(token, "copy "))
		target_ptr := &TsapaObject{}
		*target_ptr = *object
		return target_ptr

	} else if token == "object" {

		return constructObject(StructType, nil)

	} else if machine.re.block.MatchString(token) {

		args_block := machine.re._args.FindString(token)
		args := strings.Split(strings.TrimRight(strings.TrimLeft(args_block, "[ "), "| "), ",")

		for i, token := range args {
			args[i] = strings.Trim(token, " ")
		}

		sentence := strings.TrimSuffix(strings.TrimPrefix(token, args_block), "]")

		obj := constructObject(CallableType, sentence)
		obj.fields["_args"] = constructObject(_args_arr, args)
		return obj

	} else if machine.re.numeric.MatchString(token) {

		val, _ := strconv.Atoi(token)
		return constructObject(NumericType, val)

	} else if machine.re.float.MatchString(token) {

		val, _ := strconv.ParseFloat(token, 64)
		return constructObject(FloatType, val)

	} else if machine.re.boolvar.MatchString(token) {

		if token == "true" {
			return constructObject(BooleanType, true)
		} else {
			return constructObject(BooleanType, false)
		}

	} else if machine.re.macro.MatchString(token) {

		return machine._processMacro(token)

	} else if machine.re.key.MatchString(token) {

		return machine._getObject(token)

	} else if machine.re.parenth.MatchString(token) {

		contents := strings.TrimPrefix(strings.TrimSuffix(token, ")"), "(")
		return machine._eval(contents, rec+1)

	} else if token == "" {

		return nullObject()

	}

	panic("no pattern")
	return nullObject() //Must never get here
}

func (machine *TsapaMachine) _processMacro(token string) *TsapaObject {
	tokens := strings.Split(token, " ")
	if tokens[0] == "!inspect" {
		var obj *TsapaObject
		var obj_name string
		if val, test := machine.variableScope[tokens[1]]; test {
			obj = val //workaroud for types not fully supported by _eval()
			obj_name = tokens[1]
		} else {
			obj = machine._eval(strings.TrimPrefix(token, "!inspect "), 0)
			obj_name = "reflected_object"
		}
		s := ""
		s += fmt.Sprintf("%s@%p: %s", typeStringMap()[obj.type_str], &obj, obj.toString())
		for name, field := range obj.fields {
			s += "\n"
			s += fmt.Sprintf("%s:%s\t%s\t(%s)", obj_name, name, field.toString(), typeStringMap()[field.type_str])
		}
		return constructObject(StringType, s)
		//panic("!inspect invocation error: no such object")
	} else if tokens[0] == "!panic" {
		panic(tokens[1])
	} else if tokens[0] == "!reflect" {
		if val, test := machine.variableScope[tokens[1]]; test {
			defer func() {
				panic("cannot reflect " + tokens[1] + " (" + tokens[1] + " is not CallableType?)")
			}()
			return constructObject(StringType, val.value().(string))
		} else {
			panic(tokens[1] + " is inexistent")
		}
	} else if tokens[0] == "!trace" {
		machine.trace = true
		machine._eval(strings.TrimPrefix(token, "!trace "), 0)
		machine.trace = false
	} else if tokens[0] == "!scope" {
		s := "In (variableScope):\n"
		for name, object := range machine.variableScope {
			s += fmt.Sprintf("%s\t%s\t%s\n", name, typeStringMap()[object.type_str], object.toString())
		}
		s += "\nIn (tempScope):\n"
		for name, object := range machine.tempScope {
			s += fmt.Sprintf("%s\t%s\t%s\n", name, typeStringMap()[object.type_str], object.toString())
		}
		s += "\nIn (builtinScope):\n"
		for name, object := range machine.builtinScope {
			s += fmt.Sprintf("%s\t%s\t%s\n", name, typeStringMap()[object.type_str], object.toString())
		}
		return constructObject(StringType, s)
	} else if tokens[0] == "!tsapa" {
		return constructObject(StringType, "This is Gravitsapa version "+Version+".\n(c) K+Z, 2017")
	} else if tokens[0] == "!exit" || tokens[0] == "!quit" {
		fmt.Println("Bye!")
		os.Exit(0)
	} else {
		panic("no such macrodefinition")
	}
	return nullObject()
}
