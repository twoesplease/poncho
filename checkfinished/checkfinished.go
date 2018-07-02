package checkfinished

import (
	"bufio"
	"github.com/fatih/color"
	"os"
	"strings"
)

func IsUserDone() string {
	checkinput := bufio.NewReader(os.Stdin)
	color.Cyan("\nWould you like more weather data, or is that it for now?")
	color.Cyan("Type \"more please\" to get more weather data, or \"exit\" to close the program.")
	stayOrExit, _ := checkinput.ReadString('\n')
	stayOrExit = strings.TrimSuffix(stayOrExit, "\n")
	return stayOrExit
}
