package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestImgUpsertInsert(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")

	date := time.Date(2022, 1, 15, 18, 0, 0, 0, util.GetDefaultLocation())

	props := make(map[string]interface{})
	props["a"] = 1

	level := make(map[string]int8)
	level["1"] = 3

	img1 := dao.Img{
		ReportId:         100,
		Source:           "TEST",
		RemoteId:         "100",
		RawUrl:           "http://test.ru/100.png",
		Url:              "http://test.ru/100.png",
		PreviewUrl:       "http://test.ru/100p.png",
		DatePublished:    time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		LabelsForSearch:  []string{"label1", "label2"},
		Enabled:          true,
		Type:             dao.IMAGE_TYPE_IMAGE,
		Date:             &date,
		DateLevelUpdated: time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		Level:            level,
		Props:            props,
	}
	img2 := dao.Img{
		ReportId:         0,
		Source:           "TEST",
		RemoteId:         "101",
		RawUrl:           "http://test.ru/101.png",
		Url:              "http://test.ru/101.png",
		PreviewUrl:       "http://test.ru/101p.png",
		DatePublished:    time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		LabelsForSearch:  []string{"label1", "label2"},
		Enabled:          true,
		Type:             dao.IMAGE_TYPE_SCHEMA,
		Date:             nil,
		DateLevelUpdated: time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		Level:            nil,
		Props:            nil,
	}
	inserted, err := imgDao.Upsert(
		img1,
		img2,
	)

	assert.Nil(t, err)

	p := make(map[string]string)
	if len(inserted) >= 2 {
		p["id1"] = strconv.Itoa(int(inserted[0].Id))
		p["id2"] = strconv.Itoa(int(inserted[1].Id))
	}
	daoTester.TestDatabase(t, "image", "test/expected/image_inserted.xml", p)
}

func TestImgUpsertUpdate(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	// Test update only date_published
	updated, err := imgDao.Upsert(dao.Img{
		ReportId:         100,
		Source:           "wwmap",
		RemoteId:         "1",
		RawUrl:           "http://test.ru/100.png",
		Url:              "http://test.ru/100.png",
		PreviewUrl:       "http://test.ru/100p.png",
		DatePublished:    time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		LabelsForSearch:  []string{"label1", "label2"},
		Enabled:          true,
		Type:             dao.IMAGE_TYPE_IMAGE,
		Date:             nil,
		DateLevelUpdated: time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		Level:            nil,
		Props:            nil,
	})

	assert.Nil(t, err)
	if len(updated) > 0 {
		assert.Equal(t, int64(1), updated[0].Id)
	} else {
		assert.Fail(t, "No rows updated!")
	}

	daoTester.TestDatabase(t, "image", "test/expected/image_updated.xml")
}

func TestImgInsertLocal(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")

	date := time.Date(2022, 1, 15, 18, 0, 0, 0, util.GetDefaultLocation())
	level := make(map[string]int8)
	level["1"] = 3

	inserted, err := imgDao.InsertLocal(
		dao.IMAGE_TYPE_IMAGE,
		"wwmap",
		time.Date(2022, 1, 16, 10, 0, 0, 0, util.GetDefaultLocation()),
		&date,
		level,
		util.ZeroDateUTC(),
	)
	assert.Nil(t, err)

	p := make(map[string]string)
	p["id"] = strconv.Itoa(int(inserted.Id))
	daoTester.TestDatabase(t, "image", "test/expected/image_inserted_local.xml", p)
}

func TestImgFind(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	img, found, err := imgDao.Find(1)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	// To fix zone comparsion
	img.DatePublished = img.DatePublished.In(time.UTC)

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, dao.Img{
		Id:               img.Id,
		ReportId:         0,
		Source:           "wwmap",
		RemoteId:         "1",
		RawUrl:           "",
		Url:              "",
		PreviewUrl:       "",
		DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
		LabelsForSearch:  nil,
		Enabled:          true,
		Type:             dao.IMAGE_TYPE_IMAGE,
		Date:             nil,
		DateLevelUpdated: time.Time{},
		Level:            level,
		Props:            props,
	}, img)
}

func TestImgFindNotFound(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	_, found, err := imgDao.Find(100)

	assert.Nil(t, err)
	assert.False(t, found)
}

func TestImgSetEnabledFalse(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	err := imgDao.SetEnabled(1, false)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "image", "test/expected/image_set_enabled_false.xml")
}

func TestImgSetEnabledTrue(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	err := imgDao.SetEnabled(2, true)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "image", "test/expected/image_set_enabled_true.xml")
}

func TestImgSetDateAndLevel(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	level := make(map[string]int8)
	level["0"] = 4

	err := imgDao.SetDateAndLevel(
		1,
		time.Date(2022, 01, 15, 10, 0, 0, 0, util.GetDefaultLocation()),
		level,
		time.Date(2022, 01, 16, 12, 0, 0, 0, util.GetDefaultLocation()),
	)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "image", "test/expected/image_set_date_and_level.xml")
}

func TestImgSetManualLevel(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")

	level := make(map[string]int8)
	level["0"] = 4

	levelResult, err := imgDao.SetManualLevel(
		1,
		1,
	)

	expectedResult := make(map[string]int8)
	expectedResult["1"] = 3
	expectedResult["0"] = 1

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, levelResult)
	daoTester.TestDatabase(t, "image", "test/expected/image_set_manual_level.xml")
}

func TestImgGetParentIds(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	ids, err := imgDao.GetParentIds([]int64{1, 2, 3})

	expected := make(map[int64]dao.ImageParentIds)
	expected[1] = dao.ImageParentIds{SpotId: 1}
	expected[2] = dao.ImageParentIds{SpotId: 1}
	expected[3] = dao.ImageParentIds{SpotId: 1}

	assert.Nil(t, err)
	assert.Equal(t, expected, ids)
}

func TestImgSetMain(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	err := imgDao.SetMain(1, 1)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "white_water_rapid_image", "test/expected/white_water_rapid_image_set_main.xml")
}

func TestImgSetMainTwice(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/expected/white_water_rapid_image_set_main.xml")

	err := imgDao.SetMain(1, 2)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "white_water_rapid_image", "test/expected/white_water_rapid_image_set_main_another.xml")
}

func TestImgDropMainForSpot(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/expected/white_water_rapid_image_set_main.xml")

	err := imgDao.DropMainForSpot(1)

	assert.Nil(t, err)

	daoTester.TestDatabase(t, "white_water_rapid_image", "test/white_water_rapid_image.xml")
}

func TestImgGetMainForSpot(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/expected/white_water_rapid_image_set_main.xml")

	img, found, err := imgDao.GetMainForSpot(1)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	// To fix zone comparsion
	img.DatePublished = img.DatePublished.In(time.UTC)

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, dao.Img{
		Id:               img.Id,
		ReportId:         0,
		Source:           "wwmap",
		RemoteId:         "1",
		RawUrl:           "",
		Url:              "",
		PreviewUrl:       "",
		DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
		LabelsForSearch:  nil,
		Enabled:          true,
		Type:             dao.IMAGE_TYPE_IMAGE,
		Date:             nil,
		DateLevelUpdated: time.Time{},
		Level:            level,
		Props:            props,
	}, img)
}

func TestImgRemoveBySpot(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	err := imgDao.RemoveBySpot(1, nil)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "image", "test/expected/empty.xml")
	daoTester.TestDatabase(t, "white_water_rapid_image", "test/expected/empty.xml")
}

func TestImgRemoveByRiver(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	err := imgDao.RemoveByRiver(1, nil)

	assert.Nil(t, err)
	daoTester.TestDatabase(t, "image", "test/expected/empty.xml")
	daoTester.TestDatabase(t, "white_water_rapid_image", "test/expected/empty.xml")
}

func TestImgList(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	list, err := imgDao.List(1, 2, dao.IMAGE_TYPE_IMAGE, false)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	assert.Nil(t, err)
	assert.Equal(t, 2, len(list))

	// To fix zone comparsion
	if len(list) == 2 {
		list[0].DatePublished = list[0].DatePublished.In(time.UTC)
		list[1].DatePublished = list[1].DatePublished.In(time.UTC)
	}

	assert.Equal(t, []dao.Img{
		{
			Id:               2,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "2",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          false,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            make(map[string]int8),
			Props:            make(map[string]interface{}),
		},
		{
			Id:               1,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "1",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          true,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            level,
			Props:            props,
		},
	}, list)
}

func TestImgListEnabledOnly(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	list, err := imgDao.List(1, 2, dao.IMAGE_TYPE_IMAGE, true)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	assert.Nil(t, err)
	assert.Equal(t, 1, len(list))

	// To fix zone comparsion
	if len(list) == 1 {
		list[0].DatePublished = list[0].DatePublished.In(time.UTC)
	}

	assert.Equal(t, []dao.Img{
		{
			Id:               1,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "1",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          true,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            level,
			Props:            props,
		},
	}, list)
}

func TestImgListAllBySpot(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	list, err := imgDao.ListAllBySpot(1)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	assert.Nil(t, err)
	assert.Equal(t, 3, len(list))

	// To fix zone comparsion
	if len(list) == 3 {
		list[0].DatePublished = list[0].DatePublished.In(time.UTC)
		list[1].DatePublished = list[1].DatePublished.In(time.UTC)
		list[2].DatePublished = list[2].DatePublished.In(time.UTC)
	}

	assert.Equal(t, []dao.Img{
		{
			Id:               1,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "1",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          true,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            level,
			Props:            props,
		},
		{
			Id:               2,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "2",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          false,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            make(map[string]int8),
			Props:            make(map[string]interface{}),
		},
		{
			Id:               3,
			ReportId:         0,
			Source:           "youtube",
			RemoteId:         "videoid",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          true,
			Type:             dao.IMAGE_TYPE_VIDEO,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            make(map[string]int8),
			Props:            make(map[string]interface{}),
		},
	}, list)
}

func TestImgListMainByRiver(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	list, err := imgDao.ListMainByRiver(1)

	level := make(map[string]int8)
	level["1"] = 3
	props := make(map[string]interface{})
	props["1"] = "a"

	assert.Nil(t, err)
	assert.Equal(t, 1, len(list))

	// To fix zone comparsion
	if len(list) == 1 {
		list[0].DatePublished = list[0].DatePublished.In(time.UTC)
	}

	assert.Equal(t, []dao.Img{
		{
			Id:               1,
			ReportId:         0,
			Source:           "wwmap",
			RemoteId:         "1",
			RawUrl:           "",
			Url:              "",
			PreviewUrl:       "",
			DatePublished:    time.Date(2021, 01, 16, 7, 0, 0, 0, time.UTC),
			LabelsForSearch:  nil,
			Enabled:          true,
			Type:             dao.IMAGE_TYPE_IMAGE,
			Date:             nil,
			DateLevelUpdated: time.Time{},
			Level:            level,
			Props:            props,
		},
	}, list)
}
