#!/bin/bash
set -e

echo "=== Fuzzing dice expression parser ==="
go test -fuzz=FuzzDiceExpression -fuzztime=10s ./dice/

echo "=== Fuzzing character import ==="
go test -fuzz=FuzzCharacterImport -fuzztime=10s ./handlers/

echo "=== Fuzzing compendium search ==="
go test -fuzz=FuzzCompendiumSearch -fuzztime=10s ./handlers/

echo "=== Fuzzing encounter CR parser ==="
go test -fuzz=FuzzEncounterCR -fuzztime=10s ./handlers/

echo "=== All fuzz targets completed ==="
