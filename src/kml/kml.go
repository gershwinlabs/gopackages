// The kml package is used for rendering Google Earth Keyhole Markup Language
// (KML) files.  Some of the terms used in this library are pulled from the
// KML specification (recommended reading/reference).
package kml

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type renderable interface {
	render() string
}

// KML represents the top-level KML document object.
type KML struct {
	folders []*Folder
}

// NewKML returns a pointer to a KML struct.
func NewKML() *KML {
	f := make([]*Folder, 0, 2)
	return &KML{f}
}

// AddFolder adds a new Folder to the KML document.
func (k *KML) AddFolder(folder *Folder) {
	if folder != nil {
		k.folders = append(k.folders, folder)
	}
}

// Renders the entire KML document.
func (k *KML) Render() string {
	ret := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<kml xmlns=\"http://www.opengis.net/kml/2.2\">\n"

	for _, folder := range k.folders {
		ret += folder.render()
	}

	ret += "</kml>\n"

	return ret
}

// Folder represents a folder in the KML document.
type Folder struct {
	name        string
	description string
	features    []renderable
}

// Returns a pointer to a new Folder instance.
func NewFolder(name string, desc string) *Folder {
	f := make([]renderable, 0, 10)
	return &Folder{name, desc, f}
}

// AddFeature adds a feature (Placemark, another Folder, etc.) to
// the Folder.
func (f *Folder) AddFeature(feature renderable) {
	if feature != nil {
		f.features = append(f.features, feature)
	}
}

func (f *Folder) render() string {
	ret := "<Folder>\n" +
		fmt.Sprintf("<name>%s</name>\n", f.name) +
		fmt.Sprintf("<description>%s</description>\n", f.description)

	for _, feature := range f.features {
		ret += feature.render()
	}

	ret += "</Folder>\n"

	return ret
}

// Style represents a style used for a geometry object (point, line,
// polygon, etc.)
type Style struct {
	name      string
	alpha     uint8
	red       uint8
	green     uint8
	blue      uint8
	iconURL   string
	iconScale float64
	fill      int8
}

// NewStyle returns a new instance of a Style.  The alpha, red, green, and
// blue color properties are applied to point icon color as well as line and
// polygon color.  Name must be a single word (no spaces).
func NewStyle(name string, alpha uint8, red uint8, green uint8, blue uint8) *Style {
	return &Style{name, alpha, red, green, blue, "http://maps.google.com/mapfiles/kml/pushpin/ylw-pushpin.png", 1.1, 1}
}

// SetIconURL changes the icon that will be used for point placemarks.
// Built-in icon URL's can be seen in Google Earth when setting the
// placemark icon in the placemark properties dialog box.
func (s *Style) SetIconURL(url string) {
	url = strings.TrimSpace(url)

	if len(url) > 0 {
		s.iconURL = url
	}
}

// SetIconScale changes the icon scale from the default of 1.1.  Valid values
// are between 0.0 and 100.0.  Invalid values are ignored.
func (s *Style) SetIconScale(scale float64) {
	if scale >= 0.0 && scale <= 100.0 {
		s.iconScale = scale
	}
}

// SetPolygonFill specifies whether to fill in polygons.  The default is to
// not fill in the polygon.
func (s *Style) SetPolygonFill(fill bool) {
	if fill == true {
		s.fill = 1
	} else {
		s.fill = 0
	}
}

func (s *Style) render() string {
	colorStr := fmt.Sprintf("<color>%02x%02x%02x%02x</color>\n", s.alpha, s.blue, s.green, s.red) // yes, ABGR
	ret := fmt.Sprintf("<Style id=\"%s\">\n", s.name) +
		"<IconStyle>\n" +
		colorStr +
		fmt.Sprintf("<scale>%f</scale>\n", s.iconScale) +
		fmt.Sprintf("<Icon><href>%s</href></Icon>\n", s.iconURL) +
		"</IconStyle>\n" +
		"<LineStyle>\n" +
		colorStr +
		"<width>3</width>\n" +
		"</LineStyle>\n" +
		"<PolyStyle>\n" +
		colorStr +
		"<colorMode>normal</colorMode>\n" +
		fmt.Sprintf("<fill>%d</fill>\n", s.fill) +
		"<outline>1</outline>\n" +
		"</PolyStyle>\n" +
		"</Style>\n"

	return ret
}

// Point represents a point on the Earth
type Point struct {
	Lat float64 // latitude
	Lon float64 // longitude
	Alt float64 // altitude in meters
}

// NewPoint returns a pointer to a new Point instance.  An error is returned
// if the latitude or longitude are invalid.
func NewPoint(lat float64, lon float64, alt float64) (*Point, error) {
	if math.IsNaN(lat) || math.IsInf(lat, 0) {
		return nil, errors.New("Lat is NaN or Inf.")
	}

	if lat > 90.0 || lat < -90.0 {
		return nil, errors.New(fmt.Sprintf("Invalid Lat: %f", lat))
	}

	if math.IsNaN(lon) || math.IsInf(lon, 0) {
		return nil, errors.New("Lon is NaN or Inf.")
	}

	if lon > 180.0 || lon < -180.0 {
		return nil, errors.New(fmt.Sprintf("Invalid Lon: %f", lon))
	}

	if math.IsNaN(alt) || math.IsInf(alt, 0) {
		alt = 0.0
	}

	return &Point{lat, lon, alt}, nil
}

func (p *Point) render() string {
	ret := "<Point>\n" +
		"<extrude>0</extrude>\n" +
		"<altitudeMode>clampToGround</altitudeMode>\n" +
		fmt.Sprintf("<coordinates>%f,%f,%f</coordinates>\n", p.Lon, p.Lat, p.Alt) +
		"</Point>\n"

	return ret
}

// Placemark represents a placemark in the KML document.
type Placemark struct {
	name        string
	description string
	geometry    renderable
	style       string
}

// NewPlacemark returns a pointer to a new Placemark instance.  It takes a
// name, description, and a geometry object (Point, Polygon, etc.) as
// parameters.
func NewPlacemark(name string, desc string, geom renderable) *Placemark {
	return &Placemark{name, desc, geom, ""}
}

// SetStyle sets the style of the Placemark to the specified name.  The KML
// document must have a Style instance with a matching name (see NewStyle).
func (pm *Placemark) SetStyle(name string) {
	name = strings.TrimSpace(name)

	if len(name) > 0 {
		pm.style = name
	}
}

func (pm *Placemark) render() string {
	ret := "<Placemark>\n" +
		fmt.Sprintf("<name>%s</name>\n", pm.name) +
		fmt.Sprintf("<description>%s</description>\n", pm.description) +
		"<visibility>1</visibility>\n"

	if len(pm.style) > 0 {
		ret += fmt.Sprintf("<styleUrl>#%s</styleUrl>\n", pm.style)
	}

	ret += pm.geometry.render() +
		"</Placemark>\n"

	return ret
}
