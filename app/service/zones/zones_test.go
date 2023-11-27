package zones

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadKML(t *testing.T) {
	kmlBlob := `
	   	<?xml version="1.0" encoding="UTF-8"?>
	   	<kml xmlns="http://www.opengis.net/kml/2.2" xmlns:gx="http://www.google.com/kml/ext/2.2" xmlns:kml="http://www.opengis.net/kml/2.2" xmlns:atom="http://www.w3.org/2005/Atom">
	   	<Document>
	   		<name>Dead cow gulch.kml</name>
	   		<Placemark>
	   			<name>Dead cow gulch</name>
	   			<Polygon>
	   				<outerBoundaryIs>
	   					<LinearRing>
	   						<coordinates>
	   							152.6423695887568,-31.45667159676927,0 152.6422433285673,-31.45683042954698,0 152.6424731242078,-31.45701572965838,0 152.6426224066119,-31.45684082392744,0 152.6423695887568,-31.45667159676927,0
	   						</coordinates>
	   					</LinearRing>
	   				</outerBoundaryIs>
	   			</Polygon>
	   		</Placemark>
	   	</Document>
	   	</kml>`

	zone, err := UnmarkshallKML(kmlBlob)
	require.Nil(t, err)
	require.NotNil(t, zone)
	require.Equal(t, "Dead cow gulch", zone.Name)
	require.Equal(t, 5, len(zone.Polygon.Points))
	require.Equal(t, 152.6423695887568, zone.Polygon.Points[0].Longitude)
	require.Equal(t, -31.45667159676927, zone.Polygon.Points[0].Latitude)
	// good enough.
}

func TestReadKMLFile(t *testing.T) {
	zone, err := ReadKMLFile("testzones/Dead cow gulch.kml")
	require.Nil(t, err)
	require.NotNil(t, zone)
	require.Equal(t, "Dead cow gulch", zone.Name)
	require.Equal(t, 5, len(zone.Polygon.Points))
	require.Equal(t, 152.6423695887568, zone.Polygon.Points[0].Longitude)
	require.Equal(t, -31.45667159676927, zone.Polygon.Points[0].Latitude)
	// good enough.
}

func TestReadKMLDir(t *testing.T) {
	zones, err := ReadKMLDir("testzones")
	require.Nil(t, err)
	require.NotNil(t, zones)

	require.Equal(t, 2, len(zones))

	require.Equal(t, "Dead cow gulch", zones[0].Name)
	require.Equal(t, 5, len(zones[0].Polygon.Points))
	require.Equal(t, 152.6423695887568, zones[0].Polygon.Points[0].Longitude)
	require.Equal(t, -31.45667159676927, zones[0].Polygon.Points[0].Latitude)

	require.Equal(t, "Upper slopes", zones[1].Name)
	require.Equal(t, 5, len(zones[1].Polygon.Points))
	require.Equal(t, 152.6406784366888, zones[1].Polygon.Points[0].Longitude)
	require.Equal(t, -31.45642573808971, zones[1].Polygon.Points[0].Latitude)
	// good enough.
}
