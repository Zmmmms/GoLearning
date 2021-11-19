package main

import (
	"fmt"
	"log"

	"example.com/greetings"
)

func main(){

	log.SetPrefix("mod[greetings] >> ")
	log.SetFlags(1)

	names := []string{"Bebe1", "Bebe2", "Bebe3",}

	message, err := greetings.Hellos(names)
	if err != nil {
		// Fatal function will stop the program.
		log.Fatal(err)
	}

	// message := greetings.Hello("zms")
	fmt.Println(message)
}

