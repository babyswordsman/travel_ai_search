package detail

import "time"

type WalmartSku struct {
	ItemId           int64
	Timestamp        int64
	Aisle            string
	ParentItemId     int64
	Color            string
	MediumImage      string
	Name             string
	BrandName        string
	CategoryPath     string
	SalePrice        float64
	ShortDescription string
	LongDescription  string
}

type RecommendWalmartSkuResponse struct {
	ProductId      string  `json:"item_id"`
	Score          float64 `json:"score"`
	Reason         string  `json:"reason"`
	ProductMainPic string  `json:"product_main_pic"`
	ProductName    string  `json:"product_name"`
	ProductPrice   float64 `json:"product_price"`
	Aisle          string  `json:"aisle"`
}

type WalmartSkuResp struct {
	WalmartSku
	Score float64
	Id    string
}

func EsToWalmartSku(itemMap map[string]any) WalmartSkuResp {
	itemId := int64(itemMap["ItemId"].(float64))
	//parentItemId := int64(itemMap["parentItemId"].(float64))
	parentItemIdV, ok := itemMap["ParentItemId"]
	parentItemId := int64(0)
	if ok {
		parentItemId = int64(parentItemIdV.(float64))
	}
	name := itemMap["Name"].(string)
	//color := itemMap["color"].(string)
	colorV, ok := itemMap["Color"]
	color := ""
	if ok {
		color = colorV.(string)
	}
	mediumImage := itemMap["MediumImage"].(string)
	//brandName := itemMap["brandName"].(string)

	brandNameV, ok := itemMap["BrandName"]
	brandName := ""
	if ok {
		brandName = brandNameV.(string)
	}
	categoryPath := itemMap["CategoryPath"].(string)
	//shortDescription := itemMap["shortDescription"].(string)
	shortDescriptionV, ok := itemMap["ShortDescription"]
	shortDescription := ""
	if ok {
		shortDescription = shortDescriptionV.(string)
	}
	//longDescription := itemMap["longDescription"].(string)
	longDescriptionV, ok := itemMap["LongDescription"]
	longDescription := ""
	if ok {
		longDescription = longDescriptionV.(string)
	}
	//salePrice := itemMap["salePrice"].(float64)
	salePriceV, ok := itemMap["SalePrice"]
	salePrice := 0.0
	if ok {
		salePrice = salePriceV.(float64)
	}
	timestampV, ok := itemMap["Timestamp"]
	timestamp := int64(0)

	if ok {
		timestamp = int64(timestampV.(float64))
	}

	aisle := itemMap["Aisle"].(string)

	sku := WalmartSkuResp{}

	sku.ItemId = itemId
	sku.Timestamp = timestamp
	sku.Aisle = aisle
	sku.ParentItemId = parentItemId
	sku.Color = color
	sku.MediumImage = mediumImage
	sku.Name = name
	sku.BrandName = brandName
	sku.CategoryPath = categoryPath
	sku.SalePrice = salePrice
	sku.ShortDescription = shortDescription
	sku.LongDescription = longDescription

	return sku
}

func ParseWalmartSku(itemMap map[string]any) WalmartSku {
	itemId := int64(itemMap["itemId"].(float64))
	//parentItemId := int64(itemMap["parentItemId"].(float64))
	parentItemIdV, ok := itemMap["parentItemId"]
	parentItemId := int64(0)
	if ok {
		parentItemId = int64(parentItemIdV.(float64))
	}
	name := itemMap["name"].(string)
	//color := itemMap["color"].(string)
	colorV, ok := itemMap["color"]
	color := ""
	if ok {
		color = colorV.(string)
	}
	mediumImage := itemMap["mediumImage"].(string)
	//brandName := itemMap["brandName"].(string)

	brandNameV, ok := itemMap["brandName"]
	brandName := ""
	if ok {
		brandName = brandNameV.(string)
	}
	categoryPath := itemMap["categoryPath"].(string)
	//shortDescription := itemMap["shortDescription"].(string)
	shortDescriptionV, ok := itemMap["shortDescription"]
	shortDescription := ""
	if ok {
		shortDescription = shortDescriptionV.(string)
	}
	//longDescription := itemMap["longDescription"].(string)
	longDescriptionV, ok := itemMap["longDescription"]
	longDescription := ""
	if ok {
		longDescription = longDescriptionV.(string)
	}
	//salePrice := itemMap["salePrice"].(float64)
	salePriceV, ok := itemMap["salePrice"]
	salePrice := 0.0
	if ok {
		salePrice = salePriceV.(float64)
	}

	aisleV, ok := itemMap["aisle"]
	aisle := ""
	if ok {
		aisle = aisleV.(string)
	}

	sku := WalmartSku{
		ItemId:           itemId,
		Timestamp:        time.Now().UnixMilli(),
		Aisle:            aisle,
		ParentItemId:     parentItemId,
		Color:            color,
		MediumImage:      mediumImage,
		Name:             name,
		BrandName:        brandName,
		CategoryPath:     categoryPath,
		SalePrice:        salePrice,
		ShortDescription: shortDescription,
		LongDescription:  longDescription,
	}

	return sku
}
