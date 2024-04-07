package qdrant

import (
	"testing"
	"travel_ai_search/search/conf"
)

func TestCreateCollection(t *testing.T) {
	config := &conf.Config{
		QdrantAddr: "127.0.0.1:6334",
	}
	client, err := InitVectorClient(config)
	if err != nil {
		t.Error("init vector client error:", err)
	}
	defer client.Close()

	collectionName := "test1"

	collectionNames, err := client.ListCollections()
	if err != nil {
		t.Error("list collection error:", err)
	}

	isExist := false

	for _, name := range collectionNames {
		if name == collectionName {
			isExist = true
			break
		}
	}
	if isExist {
		del, err := client.DeleteCollection(collectionName)
		if err != nil {
			t.Error("delete collection request error:", err)
		}
		if del {
			t.Log("del collection:", collectionName)
		} else {
			t.Error("delete collection:", collectionName, " failure,err:", err)
		}
	}

	err = client.CreateCollection(collectionName, 4, 2)
	if err != nil {
		t.Error("create collection error:", err)
	}

	collectionNames, err = client.ListCollections()
	if err != nil {
		t.Error("list collection error:", err)
	}

	isExist = false

	for _, name := range collectionNames {
		if name == collectionName {
			isExist = true
			break
		}
	}
	if !isExist {
		t.Errorf("create collection:%s maybe err", collectionName)
	}

	vectorRows := make([]*VectorRow, 0, 4)
	row1 := NewVectorRow(1, []float32{0.1, 0.04, 0.2, 0.007})
	row1.AppendString("title", "北京一日")
	row1.AppendDouble("score", 0.9)
	row1.AppendNum("days", 5)
	row2 := NewVectorRow(2, []float32{0.1, 0.03, 0.2, 0.007})
	row2.AppendString("title", "北京一日2")
	row2.AppendDouble("score", 0.9)
	row2.AppendNum("days", 3)
	vectorRows = append(vectorRows, row1)
	vectorRows = append(vectorRows, row2)

	err = client.AddVector(collectionName, vectorRows)
	if err != nil {
		t.Error("add vector error:", err)
	}
	query := [4]float32{0.105, 0.0305, 0.205, 0.00705}
	scoredPoints, err := client.Search(collectionName, query[:], 1, true, true)
	if err != nil {
		t.Error("search vector error:", err)
	}

	if len(scoredPoints) != 1 {
		t.Errorf("search point expected 1,but %d", len(scoredPoints))
	}
	resultPoint := scoredPoints[0]

	if resultPoint.Id.GetNum() != row2.Id {
		t.Errorf("search point expected id:%d,but %d", row2.Id, resultPoint.Id.GetNum())
	}

	if resultPoint.GetScore() < 0.1 {
		t.Errorf("search point expected score:%f", resultPoint.GetScore())
	}
}

func TestCeateCollection2(t *testing.T) {
	config := &conf.Config{
		QdrantAddr: "127.0.0.1:6334",
	}
	client, err := InitVectorClient(config)
	if err != nil {
		t.Error("init vector client error:", err)
	}
	defer client.Close()

	collectionNames, err := client.ListCollections()
	if err != nil {
		t.Error("list collection error:", err)
	}

	isExist := false

	for _, name := range collectionNames {
		if name == DETAIL_COLLECTION {
			isExist = true
			break
		}
	}
	if isExist {
		return
	}
	err = client.CreateCollection(DETAIL_COLLECTION, 768, 2)
	if err != nil {
		t.Error("create collection error:", err)
	}

}
