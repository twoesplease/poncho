package checkfinished

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func IsUserDone() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("\nWould you like more weather data, or is that it for now?")
	fmt.Println("Type \"more please\" to get more weather data, or \"exit\" to close the program.")
	stayOrExit, _ := checkinput.ReadString('\n')
	stayOrExit = strings.TrimSuffix(stayOrExit, "\n")
	return stayOrExit
}
