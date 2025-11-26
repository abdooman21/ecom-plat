package servelogic

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func PrintServerHelp() {
	fmt.Println("Possible commands:")
	fmt.Println("* pause")
	fmt.Println("* resume")
	fmt.Print("add <region> <prodID> <ProdName> <price>")
	fmt.Println("* quit")
	fmt.Println("* help")
}

func PrintClientHelp() {
	fmt.Println("Possible commands:")
	fmt.Println("* status")
	fmt.Println("* quit")
	fmt.Println("* help")
}
func ClientWelcome() (string, error) {
	fmt.Println("Welcome to the Peril client!")
	fmt.Println("Please enter your username:")
	words := GetInput()
	if len(words) == 0 {
		return "", errors.New("you must enter a username. goodbye")
	}
	username := words[0]
	fmt.Printf("Welcome, %s!\n", username)
	PrintClientHelp()
	return username, nil
}

func GetInput() []string {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func PrintQuit() {
	fmt.Println("I hate this ! (╯°□°)╯︵ ┻━┻")
}
