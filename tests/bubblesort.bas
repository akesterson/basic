10 DIM A#(5)
20 A#(0) = 5
21 A#(1) = 2
22 A#(2) = 4
23 A#(3) = 1
24 A#(4) = 3
30 CHANGED# = 0
35 FOR I# = 0 TO 3
36     PRINT I#
45     J# = I#+1
46     PRINT "CHECKING A#(" + I# + ")[" + A#(I#) + "] <= A#(" + J# + ")[" + A#(J#) + "]"  
50     IF A#(I#) <= A#(J#) THEN GOTO 100
55     PRINT "TRANSPOSING A#(" + I# + ")[" + A#(I#) + "] <- A#(" + J# + ")[" + A#(J#) + "]"
60     T# = A#(I#)
70     A#(I#) = A#(H#)
80     A#(H#) = T#
85     CHANGED# = CHANGED# + 1
100 NEXT I#
105 PRINT "CHANGED " + CHANGED# + " ELEMENTS"
110 IF CHANGED# <> 0 THEN GOTO 30
120 FOR I# = 0 TO 4
130   PRINT A#(I#)
140 NEXT I#
