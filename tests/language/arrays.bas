10 DIM A#(3)
20 A#(0) = 100
30 A#(1) = 101
40 A#(2) = 102
50 IF LEN(A#) <> 3 THEN PRINT "LEN(A#) != 3 : " + LEN(A#) + " FAILURE"
60 IF A#(0) <> 100 THEN PRINT "A#(0) != 100 : " + A#(0) + " FAILURE"
70 IF A#(1) <> 101 THEN PRINT "A#(1) != 101 : " + A#(1) + " FAILURE"
80 IF A#(2) <> 102 THEN PRINT "A#(2) != 102 : " + A#(2) + " FAILURE"
90 PRINT "SUCCESS"
