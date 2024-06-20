package kvclient

import (
	"fmt"
	"sync"
)

var GLOBAL_DETAIL_ID_KEY = "global_detail_id"
var GLOBAL_SKU_ID_KEY = "global_sku_id"

type IdGen struct {
	key          string
	cacheIdUpper uint64
	nextId       uint64
	lock         sync.Mutex
	step         uint64
}

var detailIdGen *IdGen
var skuIdGen *IdGen

func StartIdGen(){
	detailIdGen = InitIdGen(GLOBAL_DETAIL_ID_KEY)
	skuIdGen = InitIdGen(GLOBAL_SKU_ID_KEY)
}


func FetchDetailNextId() (uint64, error) {
	return detailIdGen.NextId()
}

func FetchSkuNextId() (uint64, error) {
	return skuIdGen.NextId()
}

func InitIdGen(key_ string) *IdGen {
	idGen := &IdGen{
		step:         100,
		lock:         sync.Mutex{},
		cacheIdUpper: 0,
		nextId:       0,
		key:          key_,
	}
	for i := 0; i < 3; i++ {
		val, err := GetInstance().IncrBy(idGen.key, int64(idGen.step))
		if err != nil {
			continue
		}
		if val <= 0 {
			continue
		}
		idGen.cacheIdUpper = uint64(val)
		idGen.nextId = idGen.cacheIdUpper - idGen.step + 1
		break
	}

	return idGen
}

func (idGen *IdGen) NextId() (uint64, error) {
	idGen.lock.Lock()
	defer idGen.lock.Unlock()

	if idGen.cacheIdUpper <= idGen.nextId {
		for i := 0; i < 3; i++ {
			step := max(int64(idGen.step), int64(idGen.nextId)-int64(idGen.cacheIdUpper))
			val, err := GetInstance().IncrBy(idGen.key, step)
			if err != nil {
				continue
			}
			if val <= 0 {
				continue
			}
			idGen.cacheIdUpper = uint64(val)
			break
		}
	}
	if idGen.cacheIdUpper <= idGen.nextId {
		return 0, fmt.Errorf("fetch id{%s} err,upper:%d,current:%d", idGen.key, idGen.cacheIdUpper, idGen.nextId)
	}
	current := idGen.nextId
	idGen.nextId++
	return current, nil
}
