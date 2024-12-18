#!/bin/bash

GOGC=off go test -bench=.  -v -count=1 -timeout=30s -tags bench  -memprofile=mem.prof
# go tool pprof mem.prof
# (pprof) svg > mem-usage.svg 
