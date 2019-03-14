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
*		    	|- register --target <parcelID> --custody <key>
*				|- request --target <parcelID> --payment <uint64>
*				|- cancel --target <parcelID>
*				|
*				|- grant --target <parcelID> --grantee <address> --custody <key>
*				|- revoke --target <parcelID> --grantee <address>
*				|- discard --target <parcelID>
*
*		  |- query |- balance <address>
 *				   |
*				   |- parcel <parcelID>
*				   |- request --buyer <address> --target <parcelID>
*				   |- usage --buyer <address> --target <parcelID>
*/

func main() {
	cmd.Execute()
}
