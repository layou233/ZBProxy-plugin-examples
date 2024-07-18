package main

import (
	"context"

	"github.com/layou233/zbproxy/v3"
	"github.com/layou233/zbproxy/v3/common/jsonx"
	"github.com/layou233/zbproxy/v3/common/set"
	"github.com/layou233/zbproxy/v3/config"
	"github.com/layou233/zbproxy/v3/route"
	"github.com/phuslu/log"
)

func main() {
	zbproxyConfig := &config.Root{
		Log: config.Log{
			Level: log.DebugLevel,
		},
		Services: []*config.Service{
			{
				Name:   "Hypixel-in",
				Listen: 25565,
			},
		},
		Router: config.Router{
			Rules: []*config.Rule{
				{
					Type:  "always",
					Sniff: jsonx.Listable[string]{"minecraft"},
				},
				{
					Type:      "MinecraftPlayerName",
					Parameter: jsonx.RawJSON(`""`),
					Outbound:  "Hypixel-out",
				},
				{
					Type:      "custom:HypixelGuild",
					Parameter: jsonx.RawJSON(`"Your guild name here"`),
					Outbound:  "Hypixel-out",
				},
			},
			DefaultOutbound: "RESET",
		},
		Outbounds: []*config.Outbound{
			{
				Name:          "Hypixel-out",
				TargetAddress: "mc.hypixel.net",
				TargetPort:    25565,
				Minecraft: &config.MinecraftService{
					EnableHostnameRewrite: true,
					MotdFavicon:           "{DEFAULT_MOTD}",
					MotdDescription:       "§d{NAME}§e, provided by §a§o{INFO}§r\n§c§lProxy for §6§n{HOST}:{PORT}§r",
				},
			},
		},
		Lists: map[string]set.StringSet{},
	}
	zbproxyConfig.Outbounds[0].Minecraft.OnlineCount.Online = -1

	instance, err := zbproxy.NewInstance(context.Background(), zbproxy.Options{
		Config: zbproxyConfig,
		RuleRegistry: map[string]route.CustomRuleInitializer{
			"HypixelGuild": NewHypixelGuildRule,
		},
		DisableReload: true,
	})
	if err != nil {
		panic(err)
	}

	err = instance.Start()
	if err != nil {
		panic(err)
	}

	select {} // block
	// you can replace this with some signal handling
}
