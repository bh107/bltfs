// +build ignore

package mtio

/* #include <sys/mtio.h>

const int mtioctop = MTIOCTOP;
*/
import "C"

type MTOperation C.struct_mtop
type MTGet C.struct_mtget
type MTPos C.struct_mtpos

const Sizeof_MTOperation = C.sizeof_struct_mtop
const Sizeof_MTGet = C.sizeof_struct_mtget
const Sizeof_MTPos = C.sizeof_struct_mtpos
