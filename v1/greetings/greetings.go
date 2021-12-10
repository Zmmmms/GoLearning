package greetings

import (
	"fmt"
	"errors"
	"math/rand"
	"time"
)

func Hello(name string) (string, error){
	// returns a greeting that embeds the name in a msg.

	//message := fmt.Sprintf("Hi, %v. Welcome!", name)
	if name == "" {
		return "", errors.New("empty name")
	}

	var message string
	message = fmt.Sprintf(randomFormat(), name)

	return message, nil
}

func Hellos(names []string) (map[string]string, error){
	messages := make(map[string]string)
	for _, name := range names{
		message, err := Hello(name)
		if err != nil { return nil, err}
		messages[name] = message
	}
	return messages, nil
}

func init(){
	rand.Seed( time.Now().UnixNano())
}

func randomFormat() string{
	formats := []string{
		"hi, %v",
		"see! %v",
		"Hail!, %v",
	}

	return formats[ rand.Intn(len(formats))]
}

