#!/bin/bash


echo "creating $1 files"

NUMFILES=$1

mkdir testfolder

for i in $(seq $NUMFILES); do
  echo "[$i/$NUMFILES]"
  size=$((RANDOM % 10 + 1))
  dd if=/dev/urandom of=testfolder/$(uuidgen) bs=100M count=$size
done

