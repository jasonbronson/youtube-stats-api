package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	BucketName = "YoutubeStats"
)

func getStats(g *gin.Context) {
	c := Channels{}
	err := DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		channels := strings.Split(channelList, ",")
		for _, channel := range channels {
			channelDetails := strings.Split(channel, "|")
			//log.Println(channelDetails[0])
			channelID := string(channelDetails[1])
			channelName := string(channelDetails[0])

			v := b.Get([]byte(channelID))
			fmt.Printf(" %s\n", channelName)
			r := &Response{}
			err := json.Unmarshal(v, r)
			if err != nil {
				log.Println(err)
				g.JSON(http.StatusBadRequest, gin.H{"error": err})
				return nil
			}
			c.Response = append(c.Response, *r)
		}
		return nil
	})

	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error":"No data yet."})
		return 
	}

	log.Println("get stats")
	g.JSON(http.StatusOK, c)
}

func downloadStats() {
	
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		fmt.Println("Error: error create service")
		return
	}
    
	channels := strings.Split(channelList, ",")
	for _, channel := range channels {
		channelDetails := strings.Split(channel, "|")
		//log.Println(channelDetails[0])
		channelID := string(channelDetails[1])
		channelName := string(channelDetails[0])
		
		subscribers, err := GetSubscribers(youtubeService, []string{"statistics"}, channelID)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return
		}

		videoIds, err := GetVideoID(youtubeService, []string{"id"}, channelID, *maxResults)

		var videos []Video

		for _, item := range videoIds {
			video, err := GetInfoVideo(youtubeService, []string{"statistics", "snippet", "estimatedMinutesWatched"}, item)
			if err != nil {
				continue
			}
			videos = append(videos, video)
		}
		response := Response{Subscriber: subscribers, ChannelID: channelID, ChannelName: channelName, Videos: videos}
		responseJson, _ := json.Marshal(response)
		
		if len(videos) > 0 {
			err := DB.Update(func(tx *bbolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists([]byte(BucketName))
				if err != nil {
					//return fmt.Errorf("create bucket: %s", err)
				}
				log.Println(channelID, " ", string(responseJson))
				err = b.Put([]byte(channelID), responseJson)
				return err
			})
			if err != nil {
				log.Println(err)
				return
			}
		}
		fmt.Println(string(responseJson))

	}
}


func GetSubscribers(service *youtube.Service, part []string, channelId string) (uint64, error) {
	call := service.Channels.List(part)
	call.Id(channelId)

	resp, err := call.Do()

	var subscribers uint64
	if err != nil {
		return subscribers, err
	}
	if len(resp.Items) < 1 {
		return subscribers, fmt.Errorf("dont have data in this channelId %v", channelId)
	}
	return resp.Items[0].Statistics.SubscriberCount, nil
}

func GetVideoID(service *youtube.Service, part []string, channelId string, maxResults int64) ([]string, error) {
	call := service.Search.List(part)
	call.ChannelId(channelId)
	call.MaxResults(maxResults)

	resp, err := call.Do()

	if err != nil {
		return nil, err
	}

	var ids []string
	for _, item := range resp.Items {
		ids = append(ids, item.Id.VideoId)
	}
	return ids, nil
}

func GetInfoVideo(service *youtube.Service, part []string, videoId string) (Video, error) {
	call := service.Videos.List(part)

	call.Id(videoId)
	resp, err := call.Do()

	var video Video
	if err != nil {
		return video, err
	}
	if len(resp.Items) < 1 {
		return video, fmt.Errorf("dont have data in this videoid %v", videoId)
	}
	
	video = Video{
		VideoID:  videoId,
		VideoName: resp.Items[0].Snippet.Title,
		Thumbnail: resp.Items[0].Snippet.Thumbnails.Default.Url,
		Views:    resp.Items[0].Statistics.ViewCount,
		Like:     resp.Items[0].Statistics.LikeCount,
		Dislikes: resp.Items[0].Statistics.DislikeCount,
		Comments: resp.Items[0].Statistics.CommentCount,
		//WatchHours: resp.Items[0].Statistics.
		WatchHours: 0,
	}
	return video, nil
}


type Video struct {
	VideoID  string `json:"video_id"`
	VideoName string `json:"video_name"`
	Thumbnail string `json:"thumbnail"`
	Views    uint64 `json:"views"`
	Like     uint64 `json:"likes"`
	Dislikes uint64 `json:"dislike"`
	Comments uint64 `json:"comments"`
	WatchHours float64 `json:"watch_hours"`
}

type Channels struct {
	Response []Response `json:"response"`
}

type Response struct {
	ChannelID   string  `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Subscriber uint64  `json:"subscriber"`
	Videos      []Video `json:"videos"`
}