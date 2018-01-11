package mtio

/*
#include <sys/mtio.h>

const int mtioctop = MTIOCTOP;
const int mtiocget = MTIOCGET;
const int mtiocpos = MTIOCPOS;
*/
import "C"

var (
	MTIOCTOP = C.mtioctop
	MTIOCGET = C.mtiocget
	MTIOCPOS = C.mtiocpos
)
