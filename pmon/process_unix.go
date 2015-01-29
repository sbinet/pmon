package pmon

//#include <unistd.h>
import "C"

var (
	ClockTicks = uint64(C.sysconf(C._SC_CLK_TCK))
	PageSize   = int64(C.sysconf(C._SC_PAGE_SIZE))

	clockTicksToNanosecond = (1000000000 / ClockTicks)
)
