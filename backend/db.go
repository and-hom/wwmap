package main

type Storage interface {
	getTracks() []Track
}

type DummyStorage struct {

}

func (s DummyStorage) getTracks() []Track {
	return []Track{
		Track{
			Id: int64(1),
			Title:"Track1",
			Path:[]Point{
				Point{x:55.80, y:37.30, },
				Point{x:55.80, y:37.40, },
				Point{x:55.70, y:37.30, },
				Point{x:55.70, y:37.40, },
			},
		},
		//ExtDataTrack{
		//	Title:"KML Track",
		//	FileIds:[]string{"sample"},
		//},
	}
}
