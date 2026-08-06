#!/bin/bash
set -e

echo "=== Fuzzing dice expression parser ==="
go test -fuzz=FuzzDiceExpression -fuzztime=10s ./dice/

echo "=== Fuzzing character import ==="
go test -fuzz=FuzzCharacterImport -fuzztime=3s -parallel=1 ./handlers/ # -parallel=1 + short fuzztime: DB-backed fuzzers spawn one worker per interesting input (fresh DB init each)

echo "=== Fuzzing compendium search ==="
go test -fuzz=FuzzCompendiumSearch -fuzztime=3s -parallel=1 ./handlers/

echo "=== Fuzzing encounter CR parser ==="
go test -fuzz=FuzzEncounterCR -fuzztime=3s -parallel=1 ./handlers/

echo "=== All fuzz targets completed ==="
