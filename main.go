package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const Version string = "ALPHA-0.33 build 0036"
const Prompt string = "§ "

func main() {
	fmt.Printf("This is Gravitsapa v.%s\n", Version)

	machine := constructMachine()
	reader := bufio.NewReader(os.Stdin)

	//termLaunchHandler()

	for {
		fmt.Print(Prompt)

		input, _ := reader.ReadString('\n')
		input = strings.TrimRight(input, "\r\n")

		s := machine.process(input)
		fmt.Println(s)
	}
}
