/*
	Mewwoof Name Service plugin for CoreDNS

	@author: Johnson Liu, GPT-3.5
*/

package mewwoof

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

/*
	Mewwoof plugin definition
*/

func setup(c *caddy.Controller) error {

	mewwoof_plugin := &MewwoofPlugin{}

	c.Next() // 'mewwoof'

	for c.NextBlock() { // mewwoof {...}
		switch c.Val() {
		case "controller":
			if !c.NextArg() {
				return plugin.Error(mewwoof_plugin.Name(), c.ArgErr())
			}
			mewwoof_plugin.ControllerURL = c.Val()
		default:
			return plugin.Error(mewwoof_plugin.Name(), c.ArgErr())
		}

	}

	// Initialize the plugin
	err := mewwoof_plugin.Init()
	if err != nil {
		return plugin.Error(mewwoof_plugin.Name(), fmt.Errorf("failed to init: %v", err))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return mewwoof_plugin
	})

	return nil
}

/*
	CoreDNS hook
*/

func init() {
	plugin.Register("mewwoof", setup)
}
