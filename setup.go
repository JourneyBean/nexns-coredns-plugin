/*
	Nexns CoreDNS plugin for CoreDNS

	@author: Johnson Liu, GPT-3.5
*/

package nexns

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

/*
	Nexns plugin definition
*/

func setup(c *caddy.Controller) error {

	nexns_plugin := &NexnsPlugin{}

	c.Next() // 'nexns'

	for c.NextBlock() { // nexns {...}
		switch c.Val() {
		case "controller":
			if !c.NextArg() {
				return plugin.Error(nexns_plugin.Name(), c.ArgErr())
			}

			config_url := c.Val()
			if config_url[len(config_url)-1] != '/' {
				config_url = config_url + "/"
			}
			nexns_plugin.ControllerURL = config_url

		default:
			return plugin.Error(nexns_plugin.Name(), c.ArgErr())
		}

	}

	// Initialize the plugin
	err := nexns_plugin.Init()
	if err != nil {
		return plugin.Error(nexns_plugin.Name(), fmt.Errorf("failed to init: %v", err))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		nexns_plugin.Next = next
		return nexns_plugin
	})

	return nil
}

/*
	CoreDNS hook
*/

func init() {
	plugin.Register("nexns", setup)
}
