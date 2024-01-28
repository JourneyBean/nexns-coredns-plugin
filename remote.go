package nexns

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func (p *NexnsPlugin) RequestWithCredentials(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Add auth headers
	req.Header.Add("X-CLIENT-ID", p.ClientId)
	req.Header.Add("X-CLIENT-SECRET", p.ClientSecret)

	// Do request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	// defer response.Body.Close()

	return response, nil
}

func (p *NexnsPlugin) loadAllDataFromURL() error {

	log.Println("[Nexns] Pulling all data from server.")

	// Create a request
	response, err := p.RequestWithCredentials(p.ControllerURL + "api/v1/domain/dump/")
	if err != nil {
		return fmt.Errorf("HTTP request error: %v", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Read response body error: %v", err)
	}

	domainDataList := make([]DomainData, 0)

	// Parse JSON data
	err = json.Unmarshal(body, &domainDataList)
	if err != nil {
		return fmt.Errorf("JSON parsing error: %v", err)
	}

	p.Database = *BuildTrie(domainDataList)

	log.Println("[Nexns] Successfully pulled all data from server.")

	return nil
}

func (p *NexnsPlugin) loadDomainDataFromURL(domainId int) error {

	log.Println("[Nexns] Loading domain id:", domainId)

	// Send HTTP GET request
	response, err := p.RequestWithCredentials(p.ControllerURL + "api/v1/domain/" + strconv.Itoa(domainId) + "/dump/")
	if err != nil {
		return fmt.Errorf("HTTP request error: %v", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Read response body error: %v", err)
	}

	domainData := &DomainData{}

	// Parse JSON data
	err = json.Unmarshal(body, &domainData)
	if err != nil {
		return fmt.Errorf("JSON parsing error: %v", err)
	}

	p.Database.Insert(domainData)

	log.Println("[Nexns] Successfully loaded domain id:", domainId)

	return nil
}

func (p *NexnsPlugin) connectToNotificationChannel() error {

	log.Println("[Nexns] Connecting to notification channel.")

	controllerURL := strings.Replace(p.ControllerURL, "http", "ws", 1)
	headers := http.Header{}
	headers.Add("X-CLIENT-ID", p.ClientId)
	headers.Add("X-CLIENT-SECRET", p.ClientSecret)
	conn, _, err := websocket.DefaultDialer.Dial(controllerURL+"api/v1/ws/client-notify/", headers)
	if err != nil {
		log.Println("[Nexns] Failed to connect to notification channel:", err)
		return err
	}
	log.Println("[Nexns] Successfully connected to notification channel.")
	defer conn.Close()

	for {
		// 从上游服务器读取消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("[Nexns] WebSocket connection closed. Attempting to reconnect.")
			return err
		}

		notificationData := WSNotification{}
		err = json.Unmarshal(msg, &notificationData)
		if err != nil {
			log.Println("[Nexns] Error parsing notification data:", err)
		}

		p.loadDomainDataFromURL(notificationData.Domain)
	}
}
