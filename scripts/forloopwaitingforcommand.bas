10 REM This shows the waitingForCommand utility in the BasicEnvironment
11 REM We have a FOR loop here with a condition where the loop should
12 REM not execute at all. But because the checking of the conditional is
13 REM delayed until the bottom of the loop, we run the risk of the
14 REM runtime executing every line between FOR ... NEXT even though it
15 REM shouldn't. waitingForCommand prevents this from occurring
20 FOR I# = 1 TO 1
30 PRINT "waitingForCommand FAILS if this is seen"
40 NEXT I#
50 FOR I# = 1 TO 2
60 PRINT "waitingForCommand PASS if this is seen"
70 NEXT I#
80 QUIT