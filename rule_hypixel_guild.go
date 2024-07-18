package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/layou233/zbproxy/v3/adapter"
	"github.com/layou233/zbproxy/v3/common"
	"github.com/layou233/zbproxy/v3/common/set"
	"github.com/layou233/zbproxy/v3/config"
	"github.com/layou233/zbproxy/v3/route"
	"github.com/phuslu/log"
)

type mojangPlayerAPI struct {
	ID string `json:"id"`
}

type RuleHypixelGuild struct {
	access         sync.RWMutex
	logger         *log.Logger
	config         *config.Rule
	GuildName      string
	cacheGuildUUID map[[16]byte]struct{}
	lastUpdate     time.Time
}

var _ route.Rule = (*RuleHypixelGuild)(nil)

func NewHypixelGuildRule(logger *log.Logger, config *config.Rule, listMap map[string]set.StringSet) (route.Rule, error) {
	var guildName string
	err := json.Unmarshal(config.Parameter, &guildName)
	if err != nil {
		return nil, common.Cause("parse guild name: ", err)
	}
	cache, err := requestGuildMemberMap(guildName)
	if err != nil {
		return nil, err
	}
	return &RuleHypixelGuild{
		logger:         logger,
		config:         config,
		GuildName:      guildName,
		cacheGuildUUID: cache,
		lastUpdate:     time.Now(),
	}, nil
}

func (r *RuleHypixelGuild) Config() *config.Rule {
	return r.config
}

func (r *RuleHypixelGuild) Match(metadata *adapter.Metadata) bool {
	if metadata.Minecraft == nil || metadata.Minecraft.PlayerName == "" {
		return false
	}
	if metadata.Minecraft.ProtocolVersion < 759 || metadata.Minecraft.UUID == [16]byte{} {
		resp, err := http.Get("https://api.mojang.com/users/profiles/minecraft/" + metadata.Minecraft.PlayerName)
		if err != nil {
			r.logger.Debug().Str("id", metadata.ConnectionID).Err(err).Msg("Error when requesting Mojang API")
			return false
		}
		var player mojangPlayerAPI
		err = json.NewDecoder(resp.Body).Decode(&player)
		resp.Body.Close()
		if err != nil {
			r.logger.Debug().Str("id", metadata.ConnectionID).Err(err).Msg("Error when parsing Mojang response")
			return false
		}
		// decode UUID string to UUID bytes, and save to metadata
		n, err := hex.Decode(metadata.Minecraft.UUID[:], []byte(player.ID))
		if err != nil {
			r.logger.Debug().Str("id", metadata.ConnectionID).Err(err).Msg("Error when parsing Mojang response")
			return false
		}
		if n != 16 {
			r.logger.Debug().Str("id", metadata.ConnectionID).Int("len", n).Msg("Bad UUID size")
			return false
		}
	}
	r.access.RLock()
	now := time.Now()
	if now.Sub(r.lastUpdate) > 10*time.Minute {
		r.access.RUnlock()
		r.access.Lock()
		if now.Sub(r.lastUpdate) > 10*time.Minute {
			r.logger.Debug().Str("guild", r.GuildName).Msg("Updating member list")
			r.updateCache()
		}
		r.access.Unlock()
		r.access.RLock()
	}
	_, match := r.cacheGuildUUID[metadata.Minecraft.UUID]
	r.access.RUnlock()
	return match
}

func (r *RuleHypixelGuild) updateCache() {
	cache, err := requestGuildMemberMap(r.GuildName)
	if err != nil {
		r.logger.Error().Str("guild", r.GuildName).Err(err).Msg("Error when updating member list, skipped")
		return
	}
	r.cacheGuildUUID = cache
	r.lastUpdate = time.Now()
}
