local i=getRegValue("MEM", "INPUT")
local retval=i .. "+" .. i
setRegValue("MEM", "OUTPUT", retval)

