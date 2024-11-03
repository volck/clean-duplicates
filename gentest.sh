#!/bin/bash


for i in {1..200}; do
  echo $i 
  size=$((RANDOM % 10 + 1))
  dd if=/dev/urandom of=testfolder/file$i bs=10M count=$size
done

