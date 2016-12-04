package main

import "time"

type TrackList []Track;

func (t TrackList) withoutPath() TrackList {
	newList := make([]Track, len(t))
	for i:=0;i<len(t);i++ {
		newList[i] = Track{
			Id:t[i].Id,
			Title:t[i].Title,
			Points:t[i].Points,
		}
	}
	return newList
}

type Storage interface {
	getTracks() TrackList
}

type DummyStorage struct {

}

func (s DummyStorage) getTracks() TrackList {
	return []Track{
		Track{
			Id: int64(1),
			Title:"Track1",
			Path:[]Point{
				Point{x:56.2877096985583, y:37.5007003462651, },
				Point{x:56.2877096985583, y:37.5002282774785, },
				Point{x:56.2881384381127, y:37.4998420393804, },
				Point{x:56.2891388117074, y:37.4973529494145, },
				Point{x:56.2919968777318, y:37.4951642668584, },
				Point{x:56.2927351767222, y:37.4951642668584, },
				Point{x:56.2936163536233, y:37.4946063673833, },
				Point{x:56.2953310180125, y:37.4946063673833, },
				Point{x:56.2933781997074, y:37.4879544890264, },
				Point{x:56.2940450269324, y:37.4876111662724, },
				Point{x:56.2984267449926, y:37.4852937376835, },
				Point{x:56.3024746249771, y:37.481045118604, },
				Point{x:56.3051174210983, y:37.4796289122441, },
				Point{x:56.3029746274536, y:37.468084684644, },
				Point{x:56.3015222212059, y:37.4659818327763, },
				Point{x:56.3020222361459, y:37.4649518645146, },
				Point{x:56.3011650636713, y:37.4585574782231, },
				Point{x:56.3006888484319, y:37.457999578748, },
				Point{x:56.2953548322541, y:37.4449533140996, },
				Point{x:56.2928780716538, y:37.4410909331182, },
				Point{x:56.2923541209595, y:37.4321645415166, },
				Point{x:56.2900200715184, y:37.4254697478154, },
				Point{x:56.2890197209855, y:37.4243539488652, },
				Point{x:56.2865425498548, y:37.4147409117558, },
				Point{x:56.2832076429274, y:37.408646932874, },
				Point{x:56.2801583306034, y:37.4006646788457, },
				Point{x:56.2767752152185, y:37.394055715833, },
				Point{x:56.2739160111405, y:37.3896783507207, },
				Point{x:56.2730105519558, y:37.3871892607549, },
				Point{x:56.2716521951454, y:37.3839276945929, },
				Point{x:56.269025467908, y:37.3744434035162, },
				Point{x:56.2663985595555, y:37.379807821546, },
				Point{x:56.2638908871682, y:37.379979482923, },
				Point{x:56.2619323996772, y:37.3777908003668, },
				Point{x:56.2608814624543, y:37.3758596098761, },
				Point{x:56.26006935475, y:37.374400488172, },
				Point{x:56.2588511607457, y:37.3730701125006, },
				Point{x:56.2574179414771, y:37.3709672606329, },
				Point{x:56.2562713272581, y:37.3680490172247, },
				Point{x:56.2560802215348, y:37.3662465727667, },
				Point{x:56.2551246785458, y:37.3645299589972, },
				Point{x:56.2540257744953, y:37.3661607420782, },

			},
			Points:[]EventPoint{
				EventPoint{Point{x:56.2877096985583, y:37.5007003462651, },
					2,"Старт","Начало нашего маршрута", JSONTime(time.Now())},
				EventPoint{Point{x:56.26006935475, y:37.374400488172, },
					3,"Фигня какая-то", "Из леса вышел лосось", JSONTime(time.Now())},
			},
		},
	}
}


type PostgresStorage struct {

}

func (s PostgresStorage) getTracks() []Track {
	return []Track{}
}
