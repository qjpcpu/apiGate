package rr

import (
	"strings"
	"sync/atomic"
)

type Server struct {
	Host   string `toml:"server"`
	Weight int    `toml:"weight"`
}

type Cluster struct {
	servers []Server
	weights []int
	cursor  int32
}

type Services map[string]*Cluster

func NewCluster(servers ...Server) *Cluster {
	total := 0
	w := make([]int, len(servers))
	for i, serv := range servers {
		total += serv.Weight
		w[i] = serv.Weight
	}
	if total == 0 {
		for i := range w {
			w[i] = 1
		}
	}
	cluster := &Cluster{
		servers: servers,
		weights: MakeWeightsScale(w),
	}
	return cluster
}

func (cluster *Cluster) PickServer() Server {
	i := atomic.AddInt32(&cluster.cursor, 1)
	return cluster.servers[cluster.weights[int(i)%len(cluster.weights)]]
}

func NewServices() Services {
	return make(map[string]*Cluster)
}

func (s Services) AddCluster(name string, servers []Server) {
	var cluster *Cluster
	if len(servers) == 0 {
		cluster = NewCluster(Server{
			Host: name,
		})
	} else {
		cluster = NewCluster(servers...)
	}
	if _, ok := s[name]; ok {
		panic("service name conflict " + name)
	}
	s[name] = cluster
}

func (s Services) GetCluster(name string) (*Cluster, bool) {
	c, ok := s[name]
	return c, ok
}

func (s Server) Scheme() string {
	if strings.HasPrefix(s.Host, "https://") {
		return "https"
	} else {
		return "http"
	}
}

func (s Server) HostWithoutScheme() string {
	if s.Scheme() == "https" {
		return strings.TrimPrefix(s.Host, "https://")
	} else {
		return strings.TrimPrefix(s.Host, "http://")
	}
}
