package tools

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/google-kit/services"
	"github.com/nguyenvanduocit/google-kit/util"
	"google.golang.org/api/chat/v1"
)

func RegisterGChatTool(s *server.MCPServer) {
	// List spaces tool
	listSpacesTool := mcp.NewTool("gchat_list_spaces",
		mcp.WithDescription("List all available Google Chat spaces/rooms"),
	)

	// Send message tool
	sendMessageTool := mcp.NewTool("gchat_send_message",
		mcp.WithDescription("Send a message to a Google Chat space or direct message"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to send the message to (e.g. spaces/1234567890)")),
		mcp.WithString("message", mcp.Required(), mcp.Description("Text message to send")),
		mcp.WithString("thread_name", mcp.Description("Optional thread name to reply to (e.g. spaces/1234567890/threads/abcdef)")),
		mcp.WithBoolean("use_markdown", mcp.Description("Whether to format the message using markdown (default: false)")),
	)

	// List users tool
	listUsersTool := mcp.NewTool("gchat_list_users",
		mcp.WithDescription("List or search for Google Chat users"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to list members from (e.g. spaces/1234567890)")),
		mcp.WithString("query", mcp.Description("Optional search query to filter users by name")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of users to return (default: 100)")),
	)

	// List messages tool (renamed from Get messages tool)
	listMessagesTool := mcp.NewTool("gchat_list_messages",
		mcp.WithDescription("Get messages from a Google Chat space"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to get messages from (e.g. spaces/1234567890)")),
		mcp.WithNumber("page_size", mcp.Description("Maximum number of messages to return (default: 100)")),
		mcp.WithString("page_token", mcp.Description("Page token for pagination")),
	)

	s.AddTool(listSpacesTool, util.ErrorGuard(gChatListSpacesHandler))
	s.AddTool(sendMessageTool, util.ErrorGuard(gChatSendMessageHandler))
	s.AddTool(listUsersTool, util.ErrorGuard(gChatListUsersHandler))
	s.AddTool(listMessagesTool, util.ErrorGuard(gChatListMessagesHandler))
}

func gChatListSpacesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaces, err := services.DefaultGChatService().Spaces.List().Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list spaces: %v", err)), nil
	}

	result := make([]map[string]interface{}, 0)
	for _, space := range spaces.Spaces {
		spaceInfo := map[string]interface{}{
			"name":        space.Name,
			"displayName": space.DisplayName,
			"type":        space.Type,
		}
		result = append(result, spaceInfo)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal spaces: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func gChatSendMessageHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)
	message := arguments["message"].(string)
	useMarkdown, _ := arguments["use_markdown"].(bool)
	threadName, hasThread := arguments["thread_name"].(string)

	msg := &chat.Message{
		Text: message,
	}
	
	if useMarkdown {
		msg.FormattedText = message
	}

	createCall := services.DefaultGChatService().Spaces.Messages.Create(spaceName, msg)
	if hasThread && threadName != "" {
		createCall = createCall.ThreadKey(threadName)
	}

	resp, err := createCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to send message: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Message sent successfully. Message ID: %s", resp.Name)), nil
}

func gChatListUsersHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)
	query, _ := arguments["query"].(string)
	maxResults, ok := arguments["max_results"].(float64)
	if !ok {
		maxResults = 100
	}

	// Create the list members request
	listCall := services.DefaultGChatService().Spaces.Members.List(spaceName).PageSize(int64(maxResults))
	if query != "" {
		listCall = listCall.Filter(fmt.Sprintf("member.displayName:%s", query))
	}

	// Execute the request
	members, err := listCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list users: %v", err)), nil
	}

	result := make([]map[string]interface{}, 0)
	for _, member := range members.Memberships {
		if member.Member != nil {
			userInfo := map[string]interface{}{
				"name":        member.Member.Name,
				"displayName": member.Member.DisplayName,
				"type":        member.Member.Type,
			}
			result = append(result, userInfo)
		}
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal users: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func gChatListMessagesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)
	
	// Handle optional parameters
	pageSize, ok := arguments["page_size"].(float64)
	if !ok {
		pageSize = 100
	}
	
	pageToken, _ := arguments["page_token"].(string)

	// Create the list messages request
	listCall := services.DefaultGChatService().Spaces.Messages.List(spaceName).
		OrderBy("createTime desc").
		PageSize(int64(pageSize))

	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
	}

	// Execute the request
	messages, err := listCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get messages: %v", err)), nil
	}

	result := map[string]interface{}{
		"messages":      make([]map[string]interface{}, 0),
		"nextPageToken": messages.NextPageToken,
	}
	for _, msg := range messages.Messages {

		messageInfo := map[string]interface{}{
			"name":        msg.Name,
			"sender":      msg.Sender,
			"createTime":  msg.CreateTime,
			"text":        msg.Text,
			"thread":      msg.Thread,
		}

		if len(msg.Attachment) > 0 {
					attachments := make([]map[string]interface{}, 0)
		for _, attachment := range msg.Attachment {
			attachmentInfo := map[string]interface{}{
				"name":        attachment.Name,
				"contentName": attachment.ContentName,
				"contentType": attachment.ContentType,
				"source":      attachment.Source,
				"thumbnailUri": attachment.ThumbnailUri,
				"downloadUri":  attachment.DownloadUri,
			}
			attachments = append(attachments, attachmentInfo)
		}
			messageInfo["attachments"] = attachments
		}
		result["messages"] = append(result["messages"].([]map[string]interface{}), messageInfo)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal messages: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}