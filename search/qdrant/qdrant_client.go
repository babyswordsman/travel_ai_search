package qdrant

import (
	"context"
	"fmt"
	"time"
	"travel_ai_search/search/conf"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VectorClient struct {
	conn *grpc.ClientConn
}

var vectorCli *VectorClient
var DETAIL_COLLECTION string = "detail"

func GetInstance() *VectorClient {
	return vectorCli
}

type VectorRow struct {
	Id            uint64
	Vector        []float32
	StrPayload    map[string]string
	NumPayload    map[string]int64
	DoublePayload map[string]float64
}

func NewVectorRow(id uint64, vector []float32) *VectorRow {
	row := &VectorRow{
		Id:            id,
		Vector:        vector,
		StrPayload:    make(map[string]string),
		NumPayload:    make(map[string]int64),
		DoublePayload: make(map[string]float64),
	}
	return row
}

func (row *VectorRow) AppendString(field string, value string) {
	row.StrPayload[field] = value
}
func (row *VectorRow) AppendNum(field string, value int64) {
	row.NumPayload[field] = value
}
func (row *VectorRow) AppendDouble(field string, value float64) {
	row.DoublePayload[field] = value
}

func InitVectorClient(config *conf.Config) (*VectorClient, error) {
	conn, err := grpc.Dial(config.QdrantAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	var client VectorClient
	client.conn = conn
	vectorCli = &client
	return &client, nil
}

func (client *VectorClient) Close() {
	client.conn.Close()
}

func (client *VectorClient) CreateCollection(collectionName string, vectorSize int32, segmentNum uint64) error {

	var request pb.CreateCollection
	request.CollectionName = collectionName
	request.VectorsConfig = &pb.VectorsConfig{
		Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     uint64(vectorSize),
				Distance: pb.Distance_Cosine,
			},
		},
	}
	request.OptimizersConfig = &pb.OptimizersConfigDiff{
		DefaultSegmentNumber: &segmentNum,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	response, err := pb.NewCollectionsClient(client.conn).Create(ctx, &request)
	if err != nil {
		return err
	}
	if !response.GetResult() {
		return fmt.Errorf("create collection %s err,result is %v", collectionName, response.GetResult())
	} else {
		return nil
	}

}

func (client *VectorClient) ListCollections() ([]string, error) {
	request := &pb.ListCollectionsRequest{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	response, err := pb.NewCollectionsClient(client.conn).List(ctx, request)
	if err != nil {
		return make([]string, 0), err
	}

	names := make([]string, 0, len(response.GetCollections()))
	for _, c := range response.GetCollections() {
		names = append(names, c.GetName())
	}
	return names, nil
}

func (client *VectorClient) DeleteCollection(collectionName string) (bool, error) {
	request := &pb.DeleteCollection{
		CollectionName: collectionName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	response, err := pb.NewCollectionsClient(client.conn).Delete(ctx, request)
	if err != nil {
		return false, err
	}

	return response.GetResult(), err

}

func buildPoint(row *VectorRow) *pb.PointStruct {
	point := &pb.PointStruct{
		Id: &pb.PointId{
			PointIdOptions: &pb.PointId_Num{Num: row.Id},
		},
		Vectors: &pb.Vectors{
			VectorsOptions: &pb.Vectors_Vector{
				Vector: &pb.Vector{
					Data: row.Vector,
				},
			},
		},
		Payload: make(map[string]*pb.Value),
	}
	for k, v := range row.StrPayload {
		point.Payload[k] = &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: v},
		}
	}
	for k, v := range row.NumPayload {
		point.Payload[k] = &pb.Value{
			Kind: &pb.Value_IntegerValue{IntegerValue: v},
		}
	}
	for k, v := range row.DoublePayload {
		point.Payload[k] = &pb.Value{
			Kind: &pb.Value_DoubleValue{DoubleValue: v},
		}
	}
	return point
}

func (client *VectorClient) AddVector(collectionName string, rows []*VectorRow) error {
	points := make([]*pb.PointStruct, 0, len(rows))
	for _, row := range rows {
		points = append(points, buildPoint(row))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	waitUpsert := true
	request := &pb.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         points,
	}
	_, err := pb.NewPointsClient(client.conn).Upsert(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (client *VectorClient) Search(collectionName string, vector []float32, topK uint64,
	fetchVector bool, fetchPlayload bool) ([]*pb.ScoredPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	request := &pb.SearchPoints{
		CollectionName: collectionName,
		Limit:          topK,
		Vector:         vector,
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: fetchVector},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: fetchPlayload},
		},
	}
	response, err := pb.NewPointsClient(client.conn).Search(ctx, request)
	if err != nil {
		return nil, err
	}
	scoredPoints := response.GetResult()
	return scoredPoints, nil
}

func (client *VectorClient) SearchWithSpace(collectionName string, space string, vector []float32, topK uint64,
	fetchVector bool, fetchPlayload bool) ([]*pb.ScoredPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	conds := make([]*pb.Condition, 0)
	conds = append(conds, &pb.Condition{
		ConditionOneOf: &pb.Condition_Field{
			Field: &pb.FieldCondition{
				Key: "space",
				Match: &pb.Match{
					MatchValue: &pb.Match_Keyword{
						Keyword: space,
					},
				},
			},
		},
	})
	request := &pb.SearchPoints{
		CollectionName: collectionName,
		Limit:          topK,
		Vector:         vector,
		WithVectors: &pb.WithVectorsSelector{
			SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: fetchVector},
		},
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: fetchPlayload},
		},
		Filter: &pb.Filter{
			Must: conds,
		},
	}
	response, err := pb.NewPointsClient(client.conn).Search(ctx, request)
	if err != nil {
		return nil, err
	}
	scoredPoints := response.GetResult()
	return scoredPoints, nil
}
