package domain

import "math"

// NearestDriver finds the closest available driver to the given coordinates.
// Returns nil if no available drivers exist.
// Uses the Haversine formula for geographic distance calculation.
func NearestDriver(drivers []*Driver, lat, lng float64) *Driver {
	var nearest *Driver
	minDistance := math.MaxFloat64

	for _, d := range drivers {
		if d.Location == nil {
			continue
		}

		dist := haversine(lat, lng, d.Location.Latitude, d.Location.Longitude)
		if dist < minDistance {
			minDistance = dist
			nearest = d
		}
	}

	return nearest
}

// haversine calculates the distance in kilometers between two GPS coordinates.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371.0

	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
