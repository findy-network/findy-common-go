local i=getRegValue("MEM", "INPUT")
if i == "TEST" then
	setRegValue("MEM", "OUTPUT", "OK")
else
	setRegValue("MEM", "OUTPUT", "NO")
end

