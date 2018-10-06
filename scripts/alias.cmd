@echo off

REM find out which alias is called
SET aliasFor=%~n0

REM call envcli for the alias and pass all arguments
envcli run %aliasFor% %*
