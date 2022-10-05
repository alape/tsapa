package main

import (
	"fmt"
	"reflect"
)

func inspect(thing interface{}) string {
	return fmt.Sprintf("Item [%v] of type \"%s\"\n", thing, reflect.TypeOf(thing).Name())
}
