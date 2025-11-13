package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.4"

func PrintVersion() {
	fmt.Printf("Current jscrawler version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
       _                                     __           
      (_)_____ _____ _____ ____ _ _      __ / /___   _____
     / // ___// ___// ___// __  /| | /| / // // _ \ / ___/
    / /(__  )/ /__ / /   / /_/ / | |/ |/ // //  __// /    
 __/ //____/ \___//_/    \__,_/  |__/|__//_/ \___//_/     
/___/
`
	fmt.Printf("%s\n%60s\n\n", banner, "Current jscrawler version "+version)
}
