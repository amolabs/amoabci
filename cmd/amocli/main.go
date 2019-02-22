package main

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd"
)

/* Commands (expected hierarchy)
 *
 * amoconsole |- version
 *		  	  |- status
 * 		  	  |- key |- list
 *		  	   		 |- generate <nickname>
 *					 |- remove <nickname>
 *
 *		  	  |- tx |- transfer --from <address> --to <address> --amount <number>
 *		  			|- purchase --from <address> --file <hash>
 *
 *		  	  |- query |- address <address>
 */

func main() {
	cmd.Execute()
}
