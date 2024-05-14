package searchengineapi

import (
	"context"
	"reflect"
	"travel_ai_search/search/conf"

	logger "github.com/sirupsen/logrus"
)

type HybridEngine struct{
	MaxItems int
	SearchEngines []SearchEngine
}

func (engine *HybridEngine)Search(ctx context.Context, config *conf.Config, query string) ([]SearchItem, error){
	results := make([][]SearchItem,0)
	var err error = nil
	for i := range engine.SearchEngines{
		items, err1 := engine.SearchEngines[i].Search(ctx,config,query)
		if err1 != nil {
			logger.Errorf("%s search %s err:%s",reflect.TypeOf(engine.SearchEngines[i]),query,err1.Error())
			err = err1
		}else{
			results = append(results, items)
			logger.Infof("%s search %d items",reflect.TypeOf(engine.SearchEngines[i]).String(),len(items))
		}
	}
	if len(results) == 0 {
		return nil,err
	}
	return SnakeMerge[SearchItem](engine.MaxItems,results...),nil
}
