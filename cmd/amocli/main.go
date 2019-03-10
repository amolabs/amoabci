package main

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd"
)

/* Commands (expected hierarchy)
 *
 * amocli |- version
 *		  |- status
 * 		  |- key |- list
 *		  		 |- generate <nickname>
 *				 |- remove <nickname>
 *
 *		  |- tx |- transfer --to <address> --amount <uint64>
 *				|
 *		    	|- register --target <file> --custody <key>
 *				|- request --target <file> --payment <uint64>
 *				|- cancel --target <file>
 *				|
 *				|- grant --target <file> --grantee <address> --custody <key>
 *				|- revoke --target <file> --grantee <address>
 *				|- discard --target <file>
 *
 *		  |- query |- balance <address>
 */

func main() {
	cmd.Execute()
}
