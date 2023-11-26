# Polygon utils - finding out if a point is in a polygon.

Used for figuring out what map zone a coordinate is in.

## Files

    cmd/main.go - an executable just for playing with the library. Probably prefer `../zones/zones_test.go`. Build with `go build cmd/main.go` then run `./main`
    inside.py - a Python implementation of the inside function, for reference. From https://www.geeksforgeeks.org/how-to-check-if-a-given-point-lies-inside-a-polygon/
    poly.go - The Go library with IsInside
