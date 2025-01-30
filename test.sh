#!/usr/bin/env bash

failed=0
for file in tests/*bas
do
    printf "${file} ... "
    output=${file%.bas}.txt
    ./basic.exe ${file} > tmpfile
    if [[ $(md5sum tmpfile ${output} | cut -d ' ' -f 1 | sort -u | wc -l) -gt 1 ]]; then
	failed=$((failed + 1))
	echo " FAIL"
    else
	echo " PASS"
    fi
    rm -f tmpfile
done
exit $failed
