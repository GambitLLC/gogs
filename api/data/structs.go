package data

type Position struct {
	X        float64
	Y        float64
	Z        float64
	OnGround bool
}

type Rotation struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}
