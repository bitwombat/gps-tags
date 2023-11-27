package poly

// from: https://www.geeksforgeeks.org/how-to-check-if-a-given-point-lies-inside-a-polygon/

type Point struct {
	X, Y float64
}

type line struct {
	p1, p2 Point
}

// onLine checks if a point is on a line
func onLine(l line, p Point) bool {
	return p.X <= max(l.p1.X, l.p2.X) &&
		p.X >= min(l.p1.X, l.p2.X) &&
		p.Y <= max(l.p1.Y, l.p2.Y) &&
		p.Y >= min(l.p1.Y, l.p2.Y)
}

// direction calculates the turn direction of the sequence of 3 points
func direction(a, b, c Point) int {
	val := (b.Y-a.Y)*(c.X-b.X) - (b.X-a.X)*(c.Y-b.Y)
	if val == 0 {
		return 0 // Co-linear
	} else if val < 0 {
		return 2 // Anti-clockwise direction
	}
	return 1 // Clockwise direction
}

// isIntersect checks if two lines intersect
func isIntersect(l1, l2 line) bool {
	dir1 := direction(l1.p1, l1.p2, l2.p1)
	dir2 := direction(l1.p1, l1.p2, l2.p2)
	dir3 := direction(l2.p1, l2.p2, l1.p1)
	dir4 := direction(l2.p1, l2.p2, l1.p2)

	// When intersecting
	if dir1 != dir2 && dir3 != dir4 {
		return true
	}

	if dir1 == 0 && onLine(l1, l2.p1) ||
		dir2 == 0 && onLine(l1, l2.p2) ||
		dir3 == 0 && onLine(l2, l1.p1) ||
		dir4 == 0 && onLine(l2, l1.p2) {
		return true
	}

	return false
}

// isInside checks if a point is inside a polygon
func IsInside(poly []Point, p Point) bool {
	n := len(poly)
	if n < 3 {
		return false // Not a polygon
	}

	// Create a point at infinity, y is same as point p
	exline := line{p, Point{9999, p.Y}}
	count := 0
	i := 0
	for {
		// Forming a line/side from two consecutive points of poly
		side := line{poly[i], poly[(i+1)%n]}
		if isIntersect(side, exline) {
			if direction(side.p1, p, side.p2) == 0 {
				return onLine(side, p)
			}
			count++
		}
		i = (i + 1) % n
		if i == 0 {
			break
		}
	}

	// When count is odd
	return count&1 != 0
}

// Helper functions: max and min
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
