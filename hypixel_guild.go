package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/layou233/zbproxy/v3/common"
)

type HypixelGuildResponse struct {
	Success bool         `json:"success"`
	Guild   HypixelGuild `json:"guild"`
}

type HypixelGuild struct {
	Members []HypixelGuildMember `json:"members"`
}

type HypixelGuildMember struct {
	UUID string `json:"uuid"`
}

func requestGuildMemberMap(guildName string) (map[[16]byte]struct{}, error) {
	request, _ := http.NewRequest(http.MethodGet, "https://api.hypixel.net/v2/guild?name="+url.QueryEscape(guildName), nil)
	request.Header.Add("API-Key", HypixelAPIKey)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, common.Cause("request Hypixel API: ", err)
	}
	var hypixelResponse HypixelGuildResponse
	err = json.NewDecoder(resp.Body).Decode(&hypixelResponse)
	resp.Body.Close()
	if err != nil {
		return nil, common.Cause("parse Hypixel API: ", err)
	}
	return convertResponseToMemberMap(&hypixelResponse)
}

func convertResponseToMemberMap(response *HypixelGuildResponse) (map[[16]byte]struct{}, error) {
	if !response.Success {
		return nil, errors.New("hypixel request is failed")
	}
	memberMap := make(map[[16]byte]struct{}, len(response.Guild.Members))
	for _, member := range response.Guild.Members {
		var uuid [16]byte
		n, err := hex.Decode(uuid[:], []byte(member.UUID))
		if err != nil {
			return nil, common.Cause("parse member UUID: ", err)
		}
		if n != 16 {
			return nil, fmt.Errorf("bad UUID size: %d", n)
		}
		memberMap[uuid] = struct{}{}
	}
	return memberMap, nil
}
