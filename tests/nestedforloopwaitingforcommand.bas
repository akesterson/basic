10 REM This shows the waitingForCommand utility in the BasicEnvironment
11 REM when we have a nested for loop. The inner loop SHOULD execute, but
12 REM the outer loop should NOT execute. Therefore, neither loop should execute.
20 FOR I# = 1 TO 1
25 FOR J# = 2 TO 4
30 PRINT "waitingForCommand FAILS if this is seen"
35 NEXT J#
40 NEXT I#
80 QUIT