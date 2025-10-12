package main

import "fmt"

func checkType(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("type is int, value is %v\n", v)
	case string:
		fmt.Printf("type is string, value is %v\n", v)
	case bool:
		fmt.Printf("type is bool, value is %v\n", v)
	case chan int:
		fmt.Printf("type is chan int, value is %v\n", v)
	default:
		fmt.Printf("unknown type %T\n", v)
	}
}

func main() {
	checkType(1)
	checkType("hello")
	checkType(true)
	checkType(make(chan int))
}
