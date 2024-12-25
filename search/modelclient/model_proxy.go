package modelclient

import (
	"net/http"
	"travel_ai_search/search/conf"
	logger "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func EmbeddingQuery(c *gin.Context){
	queries := make([]string,0) 
	if err := c.ShouldBindBodyWith(&queries,binding.JSON);err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().QueryEmbedding(queries)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"err":conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"embs":values,
	})
}


func EmbeddingPassage(c *gin.Context){
	passages := make([]string,0) 
	if err := c.ShouldBindBodyWith(&passages,binding.JSON);err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().PassageEmbedding(passages)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"err":conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"embs":values,
	})
}

func PredictReranker(c *gin.Context){
	query_passages := make([][2]string,0) 
	if err := c.ShouldBindBodyWith(&query_passages,binding.JSON);err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().PredictorRerankerScore(query_passages)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"err":conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"scores":values,
	})
}


