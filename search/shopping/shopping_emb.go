package shopping

import (
	"context"
	"travel_ai_search/search/modelclient"
)

type ShoppingEmbdder struct {
	client *modelclient.ModelClient
}

func NewEmbedder() *ShoppingEmbdder {
	return &ShoppingEmbdder{
		client: modelclient.GetInstance(),
	}
}

// EmbedDocuments returns a vector for each text.
func (emb *ShoppingEmbdder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return emb.client.PassageEmbedding(texts)
}

// EmbedQuery embeds a single text.
func (emb *ShoppingEmbdder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	values, err := emb.client.QueryEmbedding([]string{text})
	if err != nil {
		return nil, err
	}
	return values[0], nil
}
