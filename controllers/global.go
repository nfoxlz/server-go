// global
package controllers

import (
	"server/util"
)

func errorExit() {
	if p := recover(); nil != p {
		util.LogError(p)
	}
}
