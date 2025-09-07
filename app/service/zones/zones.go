package zones

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	polypkg "github.com/bitwombat/gps-tags/poly"
)

type Zone struct {
	Name    string      `xml:"Document>Placemark>name"`
	Polygon Coordinates `xml:"Document>Placemark>Polygon>outerBoundaryIs>LinearRing>coordinates"`
}

type Coordinates struct {
	Points []Point
}

type Point struct {
	Longitude float64
	Latitude  float64
	Altitude  float64
}

func (c *Coordinates) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	err := d.DecodeElement(&value, &start)
	if err != nil {
		return fmt.Errorf("while decoding XML element: %w", err)
	}

	for _, str := range strings.Fields(value) {
		var p Point
		_, err := fmt.Sscanf(str, "%f,%f,%f", &p.Longitude, &p.Latitude, &p.Altitude)
		if err != nil {
			panic("Sscanf failed")
		}

		c.Points = append(c.Points, p)
	}
	return nil
}

func UnmarkshallKML(kmlBlob string) (Zone, error) {
	var z Zone
	err := xml.Unmarshal([]byte(kmlBlob), &z)
	if err != nil {
		return Zone{}, fmt.Errorf("while unmarshalling KML: %w", err)
	}

	return z, nil
}

func ReadKMLFile(filename string) (Zone, error) {
	kmlBlob, err := os.ReadFile(filename)
	if err != nil {
		return Zone{}, fmt.Errorf("while reading KML file: %w", err)
	}

	return UnmarkshallKML(string(kmlBlob))
}

func ReadKMLDir(path string) ([]Zone, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("while reading KML directory: %w", err)
	}

	var zones []Zone
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".kml") {
			zone, err := ReadKMLFile(path + "/" + file.Name())
			if err != nil {
				return nil, fmt.Errorf("while reading KML file: %w", err)
			}
			zones = append(zones, zone)
		}
	}

	return zones, nil
}

func (z *Zone) IsInside(p Point) bool {
	poly := make([]polypkg.Point, len(z.Polygon.Points))
	for i, point := range z.Polygon.Points {
		poly[i] = polypkg.Point{X: point.Longitude, Y: point.Latitude}
	}

	point := polypkg.Point{X: p.Longitude, Y: p.Latitude}

	return polypkg.IsInside(poly, point)
}

func NameThatZone(zones []Zone, p Point) string {
	for _, zone := range zones {
		if zone.IsInside(p) {
			return zone.Name
		}
	}

	return "Not in any known zone."
}
