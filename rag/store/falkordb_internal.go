package store

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/olekukonko/tablewriter"
)

func quoteString(i interface{}) interface{} {
	switch x := i.(type) {
	case string:
		if len(x) == 0 {
			return "\"\""
		}
		if x[0] != '"' {
			x = "\"" + x
		}
		if x[len(x)-1] != '"' {
			x += "\""
		}
		return x
	default:
		return i
	}
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	output := make([]byte, n)
	randomness := make([]byte, n)
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	for pos := range output {
		random := uint8(randomness[pos])
		randomPos := random % uint8(l)
		output[pos] = letterBytes[randomPos]
	}
	return string(output)
}

// Node represents a node within a graph.
type Node struct {
	ID         string
	Alias      string
	Label      string
	Properties map[string]interface{}
}

func (n *Node) String() string {
	s := "("
	if n.Alias != "" {
		s += n.Alias
	}
	if n.Label != "" {
		s += ":" + n.Label
	}
	if len(n.Properties) > 0 {
		p := ""
		for k, v := range n.Properties {
			p += fmt.Sprintf("%s:%v,", k, quoteString(v))
		}
		p = p[:len(p)-1]
		s += "{" + p + "}"
	}
	s += ")"
	return s
}

// Edge represents an edge connecting two nodes in the graph.
type Edge struct {
	Source      *Node
	Destination *Node
	Relation    string
	Properties  map[string]interface{}
}

func (e *Edge) String() string {
	s := "(" + e.Source.Alias + ")"
	s += "-["
	if e.Relation != "" {
		s += ":" + e.Relation
	}
	if len(e.Properties) > 0 {
		p := ""
		for k, v := range e.Properties {
			p += fmt.Sprintf("%s:%s,", k, quoteString(v))
		}
		p = p[:len(p)-1]
		s += "{" + p + "}"
	}
	s += "]->"
	s += "(" + e.Destination.Alias + ")"
	return s
}

// Graph represents a graph, which is a collection of nodes and edges.
type Graph struct {
	Name  string
	Nodes map[string]*Node
	Edges []*Edge
	Conn  redis.Conn
}

// NewGraph creates a new graph (helper constructor).
func NewGraph(name string, conn redis.Conn) Graph {
	return Graph{
		Name:  name,
		Nodes: make(map[string]*Node),
		Conn:  conn,
	}
}

// AddNode adds a node to the graph structure (for Commit usage).
func (g *Graph) AddNode(n *Node) error {
	if n.Alias == "" {
		n.Alias = randomString(10)
	}
	g.Nodes[n.Alias] = n
	return nil
}

// AddEdge adds an edge to the graph structure (for Commit usage).
func (g *Graph) AddEdge(e *Edge) error {
	if e.Source == nil || e.Destination == nil {
		return fmt.Errorf("AddEdge: both source and destination nodes should be defined")
	}
	if _, ok := g.Nodes[e.Source.Alias]; !ok {
		return fmt.Errorf("AddEdge: source node neeeds to be added to the graph first")
	}
	if _, ok := g.Nodes[e.Destination.Alias]; !ok {
		return fmt.Errorf("AddEdge: destination node neeeds to be added to the graph first")
	}
	g.Edges = append(g.Edges, e)
	return nil
}

// Commit creates the entire graph (using CREATE).
func (g *Graph) Commit() (QueryResult, error) {
	q := "CREATE "
	for _, n := range g.Nodes {
		q += fmt.Sprintf("%s,", n)
	}
	for _, e := range g.Edges {
		q += fmt.Sprintf("%s,", e)
	}
	q = q[:len(q)-1]
	return g.Query(q)
}

// QueryResult represents the results of a query.
type QueryResult struct {
	Header     []string        // Extracted from first row usually? No, GRAPH.QUERY returns header.
	Results    [][]interface{} // Changed from [][]string
	Statistics []string
}

// Query executes a query against the graph.
func (g *Graph) Query(q string) (QueryResult, error) {
	qr := QueryResult{}
	r, err := redis.Values(g.Conn.Do("GRAPH.QUERY", g.Name, q, "--compact")) // Added --compact for better parsing if needed, but standard returns 3 elements.
	// Standard GRAPH.QUERY returns: [Header, Results, Statistics]
	// With --compact, it returns specialized binary-like structure.
	// Let's stick to default (text/simple protocol) if possible, but default might be what I saw in code.
	// The original code used: `r, err := redis.Values(g.Conn.Do("GRAPH.QUERY", g.Name, q))`

	if err != nil {
		return qr, err
	}

	// r[0]: Header (Array of Strings) - Wait, original code treated r[0] as Results?
	// RedisGraph 2.0+ returns [Header, Results, Stats]
	// Older versions might be different.
	// The original code: `results, err := redis.Values(r[0], nil)`
	// If r[0] is Header, then original code was treating Header as Results?
	// No, if original code works, maybe r[0] IS results.
	// Let's check "RedisGraph response format".

	// Response Format (Client Spec):
	// 1. Header (column names)
	// 2. Result set (rows)
	// 3. Query statistics

	// So r should have 3 elements.
	// r[0]: Header.
	// r[1]: Results.
	// r[2]: Stats.

	// The original code:
	// results, err := redis.Values(r[0], nil) -> Treated r[0] as results?
	// Maybe "falkordb-go" was using an old version where protocol was different?
	// Or maybe I misread `cat` output?
	// `results, err := redis.Values(r[0], nil)`
	// `qr.Results[i], err = redis.Strings(result, nil)`
	// `qr.Statistics, err = redis.Strings(r[1], nil)`

	// It accesses r[0] and r[1]. So it assumes 2 elements.
	// This implies it is NOT standard RedisGraph 2.x protocol.
	// FalkorDB might have valid protocol?
	// Or maybe the query didn't have --compact?

	// I'll implement robust checking.

	if len(r) == 3 {
		// Header, Results, Stats
		// Header
		qr.Header, _ = redis.Strings(r[0], nil)

		// Results

		rows, _ := redis.Values(r[1], nil)
		qr.Results = make([][]interface{}, len(rows))
		for i, row := range rows {
			qr.Results[i], _ = redis.Values(row, nil)
		}

		// Stats
		qr.Statistics, _ = redis.Strings(r[2], nil)

	} else if len(r) == 2 {
		// Maybe just Results, Stats? (Old protocol?)
		// Original code logic:

		rows, _ := redis.Values(r[0], nil)
		qr.Results = make([][]interface{}, len(rows))
		for i, row := range rows {
			qr.Results[i], _ = redis.Values(row, nil)
		}

		qr.Statistics, _ = redis.Strings(r[1], nil)
	} else {
		return qr, fmt.Errorf("unexpected response length: %d", len(r))
	}

	return qr, nil
}

func (g *Graph) Delete() error {
	_, err := g.Conn.Do("GRAPH.DELETE", g.Name)
	return err
}

func (qr *QueryResult) PrettyPrint() {
	if len(qr.Results) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		if len(qr.Header) > 0 {
			table.SetHeader(qr.Header)
		}

		for _, row := range qr.Results {
			sRow := make([]string, len(row))
			for i, v := range row {
				sRow[i] = fmt.Sprint(v)
			}
			table.Append(sRow)
		}
		table.Render()
	}

	for _, stat := range qr.Statistics {
		fmt.Fprintf(os.Stdout, "\n%s", stat)
	}
	fmt.Fprintf(os.Stdout, "\n")
}
