package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func runMCPServer() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := mcpResponse{JSONRPC: "2.0", ID: req.ID}

		switch req.Method {
		case "initialize":
			resp.Result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":      map[string]interface{}{"name": "perf-audit", "version": "1.0.0"},
			}
		case "tools/list":
			resp.Result = map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "run_audit",
						"description": "Run Go benchmarks, compare against baseline. Returns Markdown report. isError=true if B/op or allocs/op regressed >10%.",
						"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
					},
					{
						"name":        "reset_baseline",
						"description": "Delete .perf-baseline.json. Next run_audit saves fresh baseline.",
						"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
					},
					{
						"name":        "get_baseline",
						"description": "Read and return current .perf-baseline.json contents.",
						"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
					},
				},
			}
		case "tools/call":
			var params struct {
				Name string `json:"name"`
			}
			json.Unmarshal(req.Params, &params)
			resp.Result = handleTool(params.Name)

		case "notifications/initialized":
			continue // no response

		default:
			resp.Error = &mcpError{Code: -32601, Message: "method not found: " + req.Method}
		}

		out, _ := json.Marshal(resp)
		fmt.Println(string(out))
	}
}

func handleTool(name string) map[string]interface{} {
	switch name {
	case "run_audit":
		report, regressed, err := runAudit()
		if err != nil {
			return toolResult("error: "+err.Error(), true)
		}
		if regressed {
			report += "\n\n**EXIT 1: Regression detected. Fix heap allocations before committing.**"
		}
		return toolResult(report, regressed)

	case "reset_baseline":
		err := os.Remove(baselineFile)
		if err != nil && !os.IsNotExist(err) {
			return toolResult("error: "+err.Error(), true)
		}
		return toolResult("Baseline reset. Next run_audit saves fresh baseline.", false)

	case "get_baseline":
		data, err := os.ReadFile(baselineFile)
		if os.IsNotExist(err) {
			return toolResult("No baseline file found.", false)
		}
		if err != nil {
			return toolResult("error: "+err.Error(), true)
		}
		return toolResult(string(data), false)

	default:
		return toolResult("unknown tool: "+name, true)
	}
}

func toolResult(text string, isError bool) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{{"type": "text", "text": text}},
		"isError": isError,
	}
}
