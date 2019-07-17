#!/bin/bash

#amocli key remove t0
amocli key remove t1
amocli key remove t2
amocli key remove t3
amocli key remove d0
amocli key remove d1
amocli key remove d2
amocli key remove d3
amocli key remove u0
amocli key remove u1
amocli key remove u2
amocli key remove u3

#amocli key generate t0 --encrypt=false
amocli key generate t1 --encrypt=false
amocli key generate t2 --encrypt=false
amocli key generate t3 --encrypt=false
amocli key generate d0 --encrypt=false
amocli key generate d1 --encrypt=false
amocli key generate d2 --encrypt=false
amocli key generate d3 --encrypt=false
amocli key generate u0 --encrypt=false
amocli key generate u1 --encrypt=false
amocli key generate u2 --encrypt=false
amocli key generate u3 --encrypt=false
