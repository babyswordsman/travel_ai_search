package modelclient

import (
	"net/http"
	"travel_ai_search/search/conf"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	logger "github.com/sirupsen/logrus"
)

func EmbeddingQuery(c *gin.Context) {
	queries := make(map[string][]string)
	if err := c.ShouldBindBodyWith(&queries, binding.JSON); err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().QueryEmbedding(queries["queries"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"embs": values,
	})
}

func EmbeddingPassage(c *gin.Context) {
	passages := make(map[string][]string)
	if err := c.ShouldBindBodyWith(&passages, binding.JSON); err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().PassageEmbedding(passages["passages"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"embs": values,
	})
}

func PredictReranker(c *gin.Context) {
	query_passages := make(map[string][][2]string)
	if err := c.ShouldBindBodyWith(&query_passages, binding.JSON); err != nil {
		logger.Errorf("parse {%s} err:%s", c.GetString(gin.BodyBytesKey), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}

	values, err := GetInstance().PredictorRerankerScore(query_passages["q_p_pairs"])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": conf.ErrHint,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"scores": values,
	})
}
