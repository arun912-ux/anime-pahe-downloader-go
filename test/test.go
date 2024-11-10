package main

import "fmt"

func main() {

	values := [2]int{1, 5}

	for i:=values[0]; i<=values[1]; i++ {
		fmt.Println(i)
	}

}