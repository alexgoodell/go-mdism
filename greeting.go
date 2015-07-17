package main

import (
	"fmt"

	"github.com/mgutz/ansi"
)

func show_greeting() {

	greeting := `

                             _______                         
.--------.--------.--------.|     __|.--.--.-----.---.-.----.
|        |        |        ||__     ||  |  |  _  |  _  |   _|
|__|__|__|__|__|__|__|__|__||_______||_____|___  |___._|__|  
                                           |_____|                   
`

	greeting = ansi.Color(greeting, "blue+bh")
	fmt.Println(greeting)
}
