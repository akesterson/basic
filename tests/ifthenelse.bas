10 A# = 1
20 IF A# == 1 THEN GOTO 30 ELSE GOTO 40
30 PRINT "A# IS 1"
35 GOTO 50
45 PRINT "A# IS NOT 1"
50 IF A# == 2 THEN GOTO 60 ELSE GOTO 80
60 PRINT "A# IS 2"
65 PRINT A#
70 GOTO 90
80 PRINT "A# IS NOT 2"
90 PRINT "DONE"