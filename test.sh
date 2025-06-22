#!/usr/bin/env bash

if [[ "$(uname -o)" == "Msys" ]]; then
	basic=./basic.exe
else
	basic=./basic
fi

failed=0
for file in $(find tests/ -iname *bas)
do
    printf "${file} ... "
    output=${file%.bas}.txt
    ${basic} ${file} > tmpfile
    if [[ $(md5sum tmpfile ${output} | cut -d ' ' -f 1 | sort -u | wc -l) -gt 1 ]]; then
	failed=$((failed + 1))
	echo " FAIL"
    else
	echo " PASS"
    fi
    rm -f tmpfile
done
exit $failed
