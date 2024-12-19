#!/bin/bash

GOGC=off go test -bench=.  -v -count=1 -timeout=30s -tags bench  -cpuprofile=cpu.prof
# go tool pprof cpu.prof
# (pprof) svg > cpu-usage.svg 
