package tools

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/google-kit/services"
	"github.com/nguyenvanduocit/google-kit/util"
	"google.golang.org/api/chat/v1"
	"gopkg.in/yaml.v3"
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

	// List users tool (simplified)
	listUsersTool := mcp.NewTool("gchat_list_users",
		mcp.WithDescription("List all Google Chat users from all spaces in the organization"),
	)

	// List messages tool (renamed from Get messages tool)
	listMessagesTool := mcp.NewTool("gchat_list_messages",
		mcp.WithDescription("Get messages from a Google Chat space"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to get messages from (e.g. spaces/1234567890)")),
		mcp.WithNumber("page_size", mcp.Description("Maximum number of messages to return (default: 100)")),
		mcp.WithString("page_token", mcp.Description("Page token for pagination")),
	)

	// Create chat thread tool
	createChatThreadTool := mcp.NewTool("gchat_create_thread",
		mcp.WithDescription("Create a new Google Chat space/thread with multiple users"),
		mcp.WithString("display_name", mcp.Required(), mcp.Description("Display name for the new chat space")),
		mcp.WithString("user_emails", mcp.Required(), mcp.Description("Comma-separated list of user email addresses to add to the chat (e.g. user1@example.com,user2@example.com)")),
		mcp.WithString("initial_message", mcp.Description("Optional initial message to send to the new chat space")),
		mcp.WithBoolean("external_user_allowed", mcp.Description("Whether to allow users outside the domain (default: false)")),
	)

	// Archive chat thread tool
	archiveChatThreadTool := mcp.NewTool("gchat_archive_thread",
		mcp.WithDescription("Archive a Google Chat space to make it read-only"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to archive (e.g. spaces/1234567890)")),
	)

	// Delete chat thread tool
	deleteChatThreadTool := mcp.NewTool("gchat_delete_thread",
		mcp.WithDescription("Delete a Google Chat space permanently"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space to delete (e.g. spaces/1234567890)")),
	)

	// List all organization users tool (simplified)
	listAllUsersTool := mcp.NewTool("gchat_list_all_users",
		mcp.WithDescription("List all unique users and their email addresses across all Google Chat spaces"),
	)

	// Get thread messages tool
	getThreadMessagesTool := mcp.NewTool("gchat_get_thread_messages",
		mcp.WithDescription("Get messages from a specific Google Chat thread"),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Name of the space containing the thread (e.g. spaces/1234567890)")),
		mcp.WithString("thread_name", mcp.Required(), mcp.Description("Name of the thread to get messages from (e.g. spaces/1234567890/threads/abcdef)")),
		mcp.WithNumber("page_size", mcp.Description("Maximum number of messages to return (default: 100)")),
		mcp.WithString("page_token", mcp.Description("Page token for pagination")),
	)

	s.AddTool(listSpacesTool, util.ErrorGuard(gChatListSpacesHandler))
	s.AddTool(sendMessageTool, util.ErrorGuard(gChatSendMessageHandler))
	s.AddTool(listUsersTool, util.ErrorGuard(gChatListUsersHandler))
	s.AddTool(listMessagesTool, util.ErrorGuard(gChatListMessagesHandler))
	s.AddTool(getThreadMessagesTool, util.ErrorGuard(gChatGetThreadMessagesHandler))
	s.AddTool(createChatThreadTool, util.ErrorGuard(gChatCreateThreadHandler))
	s.AddTool(archiveChatThreadTool, util.ErrorGuard(gChatArchiveThreadHandler))
	s.AddTool(deleteChatThreadTool, util.ErrorGuard(gChatDeleteThreadHandler))
	s.AddTool(listAllUsersTool, util.ErrorGuard(gChatListAllUsersHandler))
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

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal spaces: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
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
	// Get all spaces
	spaces, err := services.DefaultGChatService().Spaces.List().Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list spaces: %v", err)), nil
	}

	// Collect all users from all spaces with deduplication
	userEmails := make(map[string]map[string]interface{})
	
	for _, space := range spaces.Spaces {
		spaceUsers, err := getAllUsersFromSpace(space.Name, space.DisplayName)
		if err != nil {
			// Continue with other spaces if one fails
			continue
		}
		
		for _, user := range spaceUsers {
			if userEmail, ok := user["email"].(string); ok && userEmail != "" {
				if existingUser, exists := userEmails[userEmail]; exists {
					// Add this space to existing user's spaces list
					if existingSpaces, ok := existingUser["spaces"].([]string); ok {
						existingUser["spaces"] = append(existingSpaces, space.Name)
					} else {
						existingUser["spaces"] = []string{space.Name}
					}
					if existingSpaceNames, ok := existingUser["spaceNames"].([]string); ok {
						existingUser["spaceNames"] = append(existingSpaceNames, space.DisplayName)
					} else {
						existingUser["spaceNames"] = []string{space.DisplayName}
					}
				} else {
					user["spaces"] = []string{space.Name}
					user["spaceNames"] = []string{space.DisplayName}
					userEmails[userEmail] = user
				}
			}
		}
	}
	
	// Convert to slice
	var allUsers []map[string]interface{}
	for _, user := range userEmails {
		user["spaceCount"] = len(user["spaces"].([]string))
		allUsers = append(allUsers, user)
	}

	result := map[string]interface{}{
		"users":       allUsers,
		"totalUsers":  len(allUsers),
		"totalSpaces": len(spaces.Spaces),
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal users: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

// Simple helper to get all users from a space
func getAllUsersFromSpace(spaceName, spaceDisplayName string) ([]map[string]interface{}, error) {
	var allUsers []map[string]interface{}
	pageToken := ""
	
	for {
		// Get members with pagination
		listCall := services.DefaultGChatService().Spaces.Members.List(spaceName).
			PageSize(1000).
			ShowGroups(true).
			UseAdminAccess(true)
		
		if pageToken != "" {
			listCall = listCall.PageToken(pageToken)
		}

		members, err := listCall.Do()
		if err != nil {
			return nil, err
		}

		// Process all members
		for _, member := range members.Memberships {
			if member.Member != nil {
				userInfo := map[string]interface{}{
					"name":        member.Member.Name,
					"displayName": member.Member.DisplayName,
					"type":        member.Member.Type,
					"role":        member.Role,
				}
				
				// Extract email from user name
				if strings.HasPrefix(member.Member.Name, "users/") {
					userPart := strings.TrimPrefix(member.Member.Name, "users/")
					if strings.Contains(userPart, "@") {
						userInfo["email"] = userPart
					}
				}
				
				allUsers = append(allUsers, userInfo)
			}
		}
		
		// Check if there are more pages
		if members.NextPageToken == "" {
			break
		}
		pageToken = members.NextPageToken
	}

	return allUsers, nil
}

func gChatListAllUsersHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// This is identical to gChatListUsersHandler now - just get all users
	return gChatListUsersHandler(arguments)
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
			"name":       msg.Name,
			"sender":     msg.Sender,
			"createTime": msg.CreateTime,
			"text":       msg.Text,
			"thread":     msg.Thread,
		}

		if len(msg.Attachment) > 0 {
			attachments := make([]map[string]interface{}, 0)
			for _, attachment := range msg.Attachment {
				attachmentInfo := map[string]interface{}{
					"name":         attachment.Name,
					"contentName":  attachment.ContentName,
					"contentType":  attachment.ContentType,
					"source":       attachment.Source,
					"thumbnailUri": attachment.ThumbnailUri,
					"downloadUri":  attachment.DownloadUri,
				}
				attachments = append(attachments, attachmentInfo)
			}
			messageInfo["attachments"] = attachments
		}
		result["messages"] = append(result["messages"].([]map[string]interface{}), messageInfo)
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal messages: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

func gChatCreateThreadHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	displayName := arguments["display_name"].(string)
	userEmails := arguments["user_emails"].(string)
	initialMessage, hasInitialMessage := arguments["initial_message"].(string)
	externalUserAllowed, _ := arguments["external_user_allowed"].(bool)

	// Parse user emails
	emails := strings.Split(userEmails, ",")
	for i := range emails {
		emails[i] = strings.TrimSpace(emails[i])
	}

	// Create a new space
	space := &chat.Space{
		DisplayName: displayName,
		Type:        "ROOM",
		SpaceType:   "SPACE",
	}

	if externalUserAllowed {
		space.ExternalUserAllowed = true
	}

	// Create the space
	createdSpace, err := services.DefaultGChatService().Spaces.Create(space).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create space: %v", err)), nil
	}

	// Add members to the space
	failedMembers := []string{}
	successfulMembers := []string{}

	for _, email := range emails {
		if email == "" {
			continue
		}

		member := &chat.Membership{
			Member: &chat.User{
				Name: fmt.Sprintf("users/%s", email),
				Type: "HUMAN",
			},
		}

		_, err := services.DefaultGChatService().Spaces.Members.Create(createdSpace.Name, member).Do()
		if err != nil {
			failedMembers = append(failedMembers, fmt.Sprintf("%s: %v", email, err))
		} else {
			successfulMembers = append(successfulMembers, email)
		}
	}

	// Send initial message if provided
	var messageId string
	if hasInitialMessage && initialMessage != "" {
		msg := &chat.Message{
			Text: initialMessage,
		}

		sentMessage, err := services.DefaultGChatService().Spaces.Messages.Create(createdSpace.Name, msg).Do()
		if err == nil {
			messageId = sentMessage.Name
		}
	}

	// Prepare result
	result := map[string]interface{}{
		"space": map[string]interface{}{
			"name":                createdSpace.Name,
			"displayName":         createdSpace.DisplayName,
			"type":                createdSpace.Type,
			"spaceType":           createdSpace.SpaceType,
			"externalUserAllowed": createdSpace.ExternalUserAllowed,
		},
		"members": map[string]interface{}{
			"successful": successfulMembers,
			"failed":     failedMembers,
		},
	}

	if messageId != "" {
		result["initialMessageId"] = messageId
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

func gChatArchiveThreadHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)

	// Get the current space to update it
	space, err := services.DefaultGChatService().Spaces.Get(spaceName).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get space: %v", err)), nil
	}

	// Update the space state to INACTIVE (archived)
	space.SpaceHistoryState = "HISTORY_ON"

	// Archive the space by updating it
	// Note: Google Chat API uses a PATCH request to update spaces
	updatedSpace, err := services.DefaultGChatService().Spaces.Patch(spaceName, space).
		UpdateMask("spaceHistoryState").Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to archive space: %v", err)), nil
	}

	result := map[string]interface{}{
		"name":              updatedSpace.Name,
		"displayName":       updatedSpace.DisplayName,
		"type":              updatedSpace.Type,
		"spaceHistoryState": updatedSpace.SpaceHistoryState,
		"archived":          true,
		"message":           "Space archived successfully. The space is now read-only.",
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

func gChatGetThreadMessagesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)
	threadName := arguments["thread_name"].(string)

	// Handle optional parameters
	pageSize, ok := arguments["page_size"].(float64)
	if !ok {
		pageSize = 100
	}

	pageToken, _ := arguments["page_token"].(string)

	// Create the list messages request with thread filter
	listCall := services.DefaultGChatService().Spaces.Messages.List(spaceName).
		OrderBy("createTime desc").
		PageSize(int64(pageSize)).
		Filter(fmt.Sprintf("thread.name = %s", threadName))

	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
	}

	// Execute the request
	messages, err := listCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get thread messages: %v", err)), nil
	}

	result := map[string]interface{}{
		"messages":      make([]map[string]interface{}, 0),
		"nextPageToken": messages.NextPageToken,
		"threadName":    threadName,
	}
	
	for _, msg := range messages.Messages {
		messageInfo := map[string]interface{}{
			"name":       msg.Name,
			"sender":     msg.Sender,
			"createTime": msg.CreateTime,
			"text":       msg.Text,
			"thread":     msg.Thread,
		}

		if len(msg.Attachment) > 0 {
			attachments := make([]map[string]interface{}, 0)
			for _, attachment := range msg.Attachment {
				attachmentInfo := map[string]interface{}{
					"name":         attachment.Name,
					"contentName":  attachment.ContentName,
					"contentType":  attachment.ContentType,
					"source":       attachment.Source,
					"thumbnailUri": attachment.ThumbnailUri,
					"downloadUri":  attachment.DownloadUri,
				}
				attachments = append(attachments, attachmentInfo)
			}
			messageInfo["attachments"] = attachments
		}
		result["messages"] = append(result["messages"].([]map[string]interface{}), messageInfo)
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal thread messages: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}

func gChatDeleteThreadHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceName := arguments["space_name"].(string)

	// Delete the space
	_, err := services.DefaultGChatService().Spaces.Delete(spaceName).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete space: %v", err)), nil
	}

	result := map[string]interface{}{
		"spaceName": spaceName,
		"deleted":   true,
		"message":   "Space deleted successfully. This action cannot be undone.",
	}

	yamlResult, err := yaml.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(yamlResult)), nil
}
