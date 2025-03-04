package containers

import (
	"fmt"
	"net"
	"starfield/plugins/starfield/records/node"

	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var P proxy.Proxy
var Lobby proxy.RegisteredServer
var Log logr.Logger

func CreateContainer(name, template string, port int) {
	n := getNodeWithLowestInstances()
	cmd := fmt.Sprintf("docker run --name %s -d -e PAPER_VELOCITY_SECRET=%s -p %d:25565 %s", name, P.Config().Forwarding.VelocitySecret, port, template)
	_, err := n.Run(cmd)
	if err != nil {
		Log.Error(err, "CreateContainer: docker command failed", "command", cmd)
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", n.Addr(), port))
	if err != nil {
		Log.Error(err, "CreateContainer: failed to resolve TCP address", "port", port)
		return
	}
	serverInfo := proxy.NewServerInfo(name, addr)
	regServer, err := P.Register(serverInfo)
	if err != nil {
		Log.Error(err, "CreateContainer: failed to register server", "serverInfo", serverInfo)
		return
	}
	Lobby = regServer
	err = GlobalManager.AddServer(name, n.Name(), Lobby)
	if err != nil {
		Log.Error(err, "CreateContainer: failed to add server to GlobalManager", "serverName", name)
	}
}

func DeleteContainer(name string) {
	n, error := GlobalManager.GetServer(name)
	if error != nil {
		Log.Error(fmt.Errorf("node not found"), "CreateContainer: node not found")
		return
	}
	cmd := fmt.Sprintf("docker rm %s --force", name)
	_, err := n.Node.Run(cmd)
	if err != nil {
		Log.Error(err, "CreateContainer: docker command failed", "command", cmd)
		return
	}
}

func getNodeWithLowestInstances() node.Node {
	var selectedNode node.Node
	minCount := int(^uint(0) >> 1)
	nodeCounts := make(map[string]int)
	for _, server := range GlobalManager.Servers {
		nodeCounts[server.Node.Addr()]++
	}
	for _, n := range GlobalManager.Nodes {
		count := nodeCounts[n.Addr()]
		if count < minCount {
			minCount = count
			selectedNode = n
		}
	}
	return selectedNode
}
