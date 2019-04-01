package common

import (
	"fmt"
	"sort"
	"strings"
)

type Cluster struct {
	Name   string            `yaml:"name"`
	Kind   string            `yaml:"kind"`
	Params map[string]string `yaml:"parameters,omitempty"`
	Nodes  []*Node           `yaml:"nodes"`

	deployment *Deployment
}

func (c *Cluster) Deploy() {
	// create cluster first
	createArgs := []string{"cluster", "create", c.Name, "--" + c.Kind}
	for k, v := range c.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	elite(createArgs...)

	hostDict := parseHostOutput(elite("host", "ls"))
	for _, n := range c.Nodes {
		id, ok := hostDict[n.Name]
		if !ok {
			panic(fmt.Errorf("cannot find host %s", n.Name))
		}
		n.Join(id)
	}
}

// parse host out to the format of Name->ID
func parseHostOutput(output string) map[string]string {
	rows := strings.Split(output, "\n")
	res := make(map[string]string)
	for i := 1; i < len(rows); i++ {
		cols := strings.Split(rows[i], " ")
		res[cols[1]] = cols[0]
	}

	return res
}

// Sort nodes in the order of role: master > leader > worker
// assign cluster to each node
func (c *Cluster) Normalize() {
	for i := range c.Nodes {
		c.Nodes[i].cluster = c
	}

	sort.Slice(c.Nodes, func(i, j int) bool {
		return c.Nodes[i].Role == RoleManager || c.Nodes[i].Role == RoleLeader
	})
}
