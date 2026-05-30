package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/SomniSom/docker-ops/internal/locale"
)

type composeStatsRow struct {
	BlockIO  string `json:"BlockIO"`
	CPUPerc  string `json:"CPUPerc"`
	ID       string `json:"ID"`
	MemPerc  string `json:"MemPerc"`
	MemUsage string `json:"MemUsage"`
	Name     string `json:"Name"`
	NetIO    string `json:"NetIO"`
	PIDs     string `json:"PIDs"`
}

type dockerPsSizeRow struct {
	ID    string `json:"ID"`
	Names string `json:"Names"`
	Size  string `json:"Size"`
}

func statsUsesCustomFormat(args []string) bool {
	for _, a := range args {
		if a == "--format" || strings.HasPrefix(a, "--format=") {
			return true
		}
	}
	return false
}

func statsIncludeAll(args []string) bool {
	for _, a := range args {
		if a == "-a" || a == "--all" {
			return true
		}
	}
	return false
}

func parseComposeStatsJSON(data []byte) ([]composeStatsRow, error) {
	var rows []composeStatsRow
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var row composeStatsRow
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseDockerPsSizes(data []byte) map[string]string {
	sizes := make(map[string]string)
	for _, row := range parseDockerPsSizeRows(data) {
		sizes[row.ID] = row.Size
	}
	return sizes
}

func parseDockerPsSizeRows(data []byte) []dockerPsSizeRow {
	var rows []dockerPsSizeRow
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var row dockerPsSizeRow
		if err := json.Unmarshal(line, &row); err != nil {
			continue
		}
		if row.ID == "" {
			continue
		}
		row.Names = strings.TrimPrefix(row.Names, "/")
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Names < rows[j].Names })
	return rows
}

func lookupDiskSize(sizes map[string]string, id string) string {
	if id == "" {
		return "-"
	}
	if s, ok := sizes[id]; ok && s != "" {
		return s
	}
	for full, size := range sizes {
		if strings.HasPrefix(full, id) || strings.HasPrefix(id, full) {
			return size
		}
	}
	return "-"
}

func printStatsWithDisk(w io.Writer, rows []composeStatsRow, sizes map[string]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CONTAINER ID\tNAME\tCPU %\tMEM USAGE / LIMIT\tMEM %\tNET I/O\tBLOCK I/O\tDISK SIZE\tPIDS")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.ID, r.Name, r.CPUPerc, r.MemUsage, r.MemPerc, r.NetIO, r.BlockIO,
			lookupDiskSize(sizes, r.ID), r.PIDs)
	}
	_ = tw.Flush()
}

func (s *composeSession) fetchDiskSizeRows(services []string, all bool) ([]dockerPsSizeRow, error) {
	project := s.cfg.ComposeProjectName
	args := []string{"ps"}
	if all {
		args = append(args, "-a")
	}
	args = append(args, "-s", "--no-trunc", "--format", "{{json .}}",
		"--filter", "label=com.docker.compose.project="+project)
	if len(services) == 1 {
		args = append(args, "--filter", "label=com.docker.compose.service="+services[0])
	}
	out, err := s.dockerOutput(args...)
	if err != nil {
		return nil, err
	}
	return parseDockerPsSizeRows(out), nil
}

func (s *composeSession) runStatsNoStreamWithSize(args []string) error {
	services, rest := splitLeadingServiceArgs(args)
	composeArgs := append([]string{"stats", "--no-stream", "--format", "json"}, services...)
	composeArgs = append(composeArgs, filterStatsPassthroughFlags(rest)...)
	out, err := s.composeOutput(composeArgs...)
	if err != nil {
		return err
	}
	rows, err := parseComposeStatsJSON(out)
	if err != nil {
		return err
	}
	sizeRows, err := s.fetchDiskSizeRows(services, statsIncludeAll(rest))
	if err != nil {
		return err
	}
	sizes := make(map[string]string, len(sizeRows))
	for _, r := range sizeRows {
		sizes[r.ID] = r.Size
	}
	printStatsWithDisk(os.Stdout, rows, sizes)
	return nil
}

func (s *composeSession) printStatsDiskHeader(services []string, all bool) error {
	rows, err := s.fetchDiskSizeRows(services, all)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, locale.T("stats.disk_header"))
	for _, r := range rows {
		name := r.Names
		if name == "" {
			name = r.ID
			if len(name) > 12 {
				name = name[:12]
			}
		}
		fmt.Printf("  %s  %s\n", name, r.Size)
	}
	fmt.Println()
	return nil
}

func filterStatsPassthroughFlags(rest []string) []string {
	var out []string
	for _, a := range rest {
		if a == "--no-stream" || strings.HasPrefix(a, "--no-stream=") {
			continue
		}
		out = append(out, a)
	}
	return out
}
