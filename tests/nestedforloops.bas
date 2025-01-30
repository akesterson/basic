10 FOR I# = 1 TO 4
15     PRINT "OUTER : I# IS " + I#
20         FOR J# = 2 TO 3
23             PRINT "INNER : I# IS " + I#
25             PRINT "INNER : J# IS " + J#
30             PRINT "INNER : I# * J# IS " + (I# * J#)
40         NEXT J#
50     NEXT I#
60 PRINT "DONE"
70 QUIT
