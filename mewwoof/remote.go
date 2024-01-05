package mewwoof

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (p *MewwoofPlugin) loadAllDataFromURL() error {

	// Send HTTP GET request
	response, err := http.Get(p.ControllerURL + "api/v1/dump/")
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

	return nil
}

func (p *MewwoofPlugin) loadDomainDataFromURL(domainId int) error {

	// Send HTTP GET request
	response, err := http.Get(p.ControllerURL + "api/v1/dump/" + strconv.Itoa(domainId) + "/")
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

	return nil
}
