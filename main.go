package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {

	r := gin.Default()
	setupRoutes(r)
	r.Run()
}

var wsupgraders = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func RegisterClient(c *gin.Context) {

	wsupgraders.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsupgraders.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to set websocket upgrade: %+v", err)
		fmt.Println(msg)
		return
	}

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 5)
		mType, mByte, err := conn.ReadMessage()
		fmt.Println("mByte: ", string(mByte))
		fmt.Println("mType: ", mType)
		fmt.Println("err: ", err)

		if string(mByte) != "nil" {
			link := getGIF(string(mByte))
			fmt.Println("link: ", link)
			image, err := downloadImage(link)
			if err == nil {
				conn.WriteMessage(websocket.BinaryMessage, image)
			} else {
				fmt.Println("error: ", err)
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", "unable to load image")))
			}

		}

	}
	conn.Close()
}

type GIF struct {
	Results []struct {
		Media []struct {
			Gif struct {
				URL string `json:"url"`
			} `json:"gif"`
		} `json:"media"`
	} `json:"results"`
}

func setupRoutes(r *gin.Engine) {
	r.GET("/connect", RegisterClient)
}

//download image from url
func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getGIF(query string) string {

	url := "https://g.tenor.com/v1/search?q=" + query + "&key=LIVDSRZULELA&limit=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "unable to generate image"
	}
	req.Header.Add("apiKey", "0UTRbFtkMxAplrohufYco5IY74U8hOes")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "unable to generate image"
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "unable to generate image"
	}
	data := GIF{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
		return "unable to generate image"
	}
	return data.Results[0].Media[0].Gif.URL
}
