package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	rdb "github.com/benauro/kube-cdn/cdn/redis"
)

// Example CDN node mapping (you can replace this with actual logic to fetch from CDN controller)
var cdnNodes = map[string]string{
	"media1": "cdn-node-1",
	"media2": "cdn-node-2",
	"media3": "cdn-node-3",
}

type (
	getMediaResponse struct {
		MediaID string `json:"mediaID"`
		CDNNode string `json:"cdnNode"`
	}
)

func GetMedia(c *gin.Context) {
	mediaID := c.Param("mediaID")

	// Check if mediaID exists in Redis cache
	cdnNode, err := rdb.Client().Get(context.Background(), mediaID).Result()
	if err == redis.Nil {
		// Media ID not found in cache, query CDN nodes (simulate for demo purpose)
		if node, ok := cdnNodes[mediaID]; ok {
			cdnNode = node

			// Cache the result in Redis for future requests (expiry set to 1 hour)
			err := rdb.Client().Set(context.Background(), mediaID, cdnNode, time.Hour).Err()
			if err != nil {
				fmt.Println("Error caching DNS result:", err)
			}
		} else {
			cdnNode = "unknown" // Handle case where media ID is not found
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve DNS information"})
		return
	}

	c.JSON(http.StatusOK,
		&getMediaResponse{
			MediaID: mediaID,
			CDNNode: cdnNode,
		},
	)
}
