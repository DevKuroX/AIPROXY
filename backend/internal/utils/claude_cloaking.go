// ref: _ref/9router/open-sse/utils/claudeCloaking.js
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	ClaudeVersion = "2.1.92"
	CCEntrypoint  = "sdk-cli"
)

// ClaudeToolSuffix is the suffix added to client tools when forwarding to CL4ude provider
var ClaudeToolSuffix = "_ide"

// CCDefaultTools are Code Assistant's native tools kept as decoys
var CCDefaultTools = map[string]bool{
	"Task":          true,
	"TaskOutput":    true,
	"TaskStop":      true,
	"TaskCreate":    true,
	"TaskGet":       true,
	"TaskUpdate":    true,
	"TaskList":      true,
	"Bash":          true,
	"Glob":          true,
	"Grep":          true,
	"Read":          true,
	"Edit":          true,
	"Write":         true,
	"NotebookEdit":  true,
	"WebFetch":      true,
	"WebSearch":     true,
	"AskUserQuestion": true,
	"Skill":         true,
	"EnterPlanMode": true,
	"ExitPlanMode":  true,
}

// CCDecoyTools are decoy tools injected after client tools
var CCDecoyTools = []map[string]interface{}{
	{
		"name":        "Task",
		"description": "Launch a new agent that has access to the following tools: Bash, Glob, Grep, Read, Edit, Write, NotebookEdit, WebFetch, WebSearch, AskUserQuestion. When you are done, use the TaskOutput tool to get the result.",
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The task for the agent to perform",
				},
			},
			"required": []string{"prompt"},
		},
	},
	{
		"name":        "Bash",
		"description": "Executes a given bash command in a persistent shell session with optional timeout",
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The command to execute",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Clear description of what this command does",
				},
			},
			"required": []string{"command", "description"},
		},
	},
}

// GenerateBillingHeader generates billing header matching real Code Assistant format
func GenerateBillingHeader(payload interface{}) string {
	content, _ := json.Marshal(payload)
	hash := sha256.Sum256(content)
	cch := hex.EncodeToString(hash[:])[:5]
	
	// Generate random build hash (3 hex chars)
	buildHash := make([]byte, 2)
	rand.Read(buildHash)
	buildHashStr := hex.EncodeToString(buildHash)[:3]
	
	return fmt.Sprintf("x-billing-info: app_version=%s.%s; app_entrypoint=%s; cch=%s;",
		ClaudeVersion, buildHashStr, CCEntrypoint, cch)
}

// GenerateFakeUserID generates fake user ID in Code Assistant JSON format
func GenerateFakeUserID(sessionID string) string {
	deviceID := make([]byte, 32)
	rand.Read(deviceID)
	deviceIDHex := hex.EncodeToString(deviceID)
	
	accountUUID := uuid.New().String()
	
	sessionUUID := sessionID
	if sessionUUID == "" {
		sessionUUID = uuid.New().String()
	}
	
	return fmt.Sprintf(`{"device_id":"%s","account_uuid":"%s","session_id":"%s"}`,
		deviceIDHex, accountUUID, sessionUUID)
}

// CloakResult contains the cloaked body and tool name mapping
type CloakResult struct {
	Body        map[string]interface{}
	ToolNameMap map[string]string // maps suffixed name -> original name
}

// CloakCL4udeTools cloaks tools before sending to CL4ude provider (anti-ban)
func CloakCL4udeTools(body map[string]interface{}) *CloakResult {
	tools, ok := body["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		return &CloakResult{Body: body, ToolNameMap: nil}
	}

	toolNameMap := make(map[string]string)
	var clientDeclarations []map[string]interface{}

	// All client tools get renamed with suffix
	for _, t := range tools {
		tool, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := tool["name"].(string)
		suffixed := name + ClaudeToolSuffix
		toolNameMap[suffixed] = name
		
		newTool := make(map[string]interface{})
		for k, v := range tool {
			newTool[k] = v
		}
		newTool["name"] = suffixed
		clientDeclarations = append(clientDeclarations, newTool)
	}

	// Client tools first, then CC decoy tools
	var allTools []interface{}
	for _, t := range clientDeclarations {
		allTools = append(allTools, t)
	}
	for _, t := range CCDecoyTools {
		allTools = append(allTools, t)
	}

	// Rename tool_use in message history
	if messages, ok := body["messages"].([]interface{}); ok {
		var renamedMessages []interface{}
		for _, m := range messages {
			msg, ok := m.(map[string]interface{})
			if !ok {
				renamedMessages = append(renamedMessages, m)
				continue
			}
			
			content, ok := msg["content"].([]interface{})
			if !ok {
				renamedMessages = append(renamedMessages, m)
				continue
			}
			
			var renamedContent []interface{}
			for _, c := range content {
				block, ok := c.(map[string]interface{})
				if !ok {
					renamedContent = append(renamedContent, c)
					continue
				}
				
				blockType, _ := block["type"].(string)
				if blockType == "tool_use" {
					newBlock := make(map[string]interface{})
					for k, v := range block {
						newBlock[k] = v
					}
					if name, ok := block["name"].(string); ok {
						newBlock["name"] = name + ClaudeToolSuffix
					}
					renamedContent = append(renamedContent, newBlock)
				} else {
					renamedContent = append(renamedContent, c)
				}
			}
			
			newMsg := make(map[string]interface{})
			for k, v := range msg {
				newMsg[k] = v
			}
			newMsg["content"] = renamedContent
			renamedMessages = append(renamedMessages, newMsg)
		}
		
		body["messages"] = renamedMessages
	}

	body["tools"] = allTools

	return &CloakResult{Body: body, ToolNameMap: toolNameMap}
}

// UncloakToolName maps suffixed tool names back to original names
func UncloakToolName(suffixedName string, toolNameMap map[string]string) string {
	if toolNameMap == nil {
		return suffixedName
	}
	if original, ok := toolNameMap[suffixedName]; ok {
		return original
	}
	return suffixedName
}
