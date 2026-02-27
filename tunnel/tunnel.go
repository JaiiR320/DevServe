package tunnel

import (
	"os/exec"
	"strconv"
)

// Tunnel abstracts a tunneling provider (Tailscale, Cloudflare, ngrok, etc.)
type Tunnel interface {
	Serve(port int) error
	Stop(port int) error
}

// TailscaleTunnel implements Tunnel using tailscale serve.
type TailscaleTunnel struct{}

func (TailscaleTunnel) Serve(port int) error {
	portStr := strconv.Itoa(port)
	return exec.Command("tailscale", "serve", "--https", portStr, "--bg", "localhost:"+portStr).Run()
}

func (TailscaleTunnel) Stop(port int) error {
	portStr := strconv.Itoa(port)
	return exec.Command("tailscale", "serve", "--https", portStr, "off").Run()
}

// DefaultTunnel is the package-level tunnel used by Process.
var DefaultTunnel Tunnel = TailscaleTunnel{}
