package common

import (
	"runtime"
	
)


func ParseLoginCookie(){

}
func GetStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}
