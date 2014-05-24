#!/bin/sh

NUMBER_OF_KB=$1
MY_PATH="`dirname \"$0\"`"

while [ $NUMBER_OF_KB -gt 0 ]
	do
		cat $MY_PATH/one_kb.txt
		sleep 0.05
		NUMBER_OF_KB=$(( $NUMBER_OF_KB - 1 ))
done
