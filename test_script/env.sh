#!/bin/bash

eval $(amocli key list | awk '{ if ($4 == "x" && $2 != "t0") printf "%s=%s\n",$2,$5 }')

# Be careful about this. This account's balance is the source of all assets.
t0=FD037CADE0A0B3C8FB5039BAC17779E9F6E8BD8F

