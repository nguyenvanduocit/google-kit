package tools

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"encoding/base64"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/google-kit/services"
	"github.com/nguyenvanduocit/google-kit/util"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func RegisterGmailTools(s *server.MCPServer) {
    // Search tool
    searchTool := mcp.NewTool("gmail_search",
        mcp.WithDescription("Search emails in Gmail using Gmail's search syntax"),
        mcp.WithString("query", mcp.Required(), mcp.Description("Gmail search query. Follow Gmail's search syntax")),
    )
    s.AddTool(searchTool, util.ErrorGuard(gmailSearchHandler))

    // Read email tool
    readEmailTool := mcp.NewTool("gmail_read_email",
        mcp.WithDescription("Read a specific email's full content including headers and body"),
        mcp.WithString("message_id", mcp.Required(), mcp.Description("ID of the email message to read")),
        mcp.WithBoolean("include_attachments", mcp.Description("Whether to include attachment information")),
    )
    s.AddTool(readEmailTool, util.ErrorGuard(gmailReadEmailHandler))

    // Reply to email tool
    replyEmailTool := mcp.NewTool("gmail_reply_email",
        mcp.WithDescription("Reply to a specific email"),
        mcp.WithString("message_id", mcp.Required(), mcp.Description("ID of the email message to reply to")),
        mcp.WithString("reply_text", mcp.Required(), mcp.Description("Text content of the reply")),
        mcp.WithBoolean("reply_all", mcp.Description("Whether to reply to all recipients")),
    )
    s.AddTool(replyEmailTool, util.ErrorGuard(gmailReplyEmailHandler))

    // Move to spam tool
    spamTool := mcp.NewTool("gmail_move_to_spam",
        mcp.WithDescription("Move specific emails to spam folder in Gmail by message IDs"),
        mcp.WithString("message_ids", mcp.Required(), mcp.Description("Comma-separated list of message IDs to move to spam")),
    )
    s.AddTool(spamTool, util.ErrorGuard(gmailMoveToSpamHandler))

    // Add create filter tool
    createFilterTool := mcp.NewTool("gmail_create_filter",
        mcp.WithDescription("Create a Gmail filter with specified criteria and actions"),
        mcp.WithString("from", mcp.Description("Filter emails from this sender")),
        mcp.WithString("to", mcp.Description("Filter emails to this recipient")),
        mcp.WithString("subject", mcp.Description("Filter emails with this subject")),
        mcp.WithString("query", mcp.Description("Additional search query criteria")),
        mcp.WithBoolean("add_label", mcp.Description("Add label to matching messages")),
        mcp.WithString("label_name", mcp.Description("Name of the label to add (required if add_label is true)")),
        mcp.WithBoolean("mark_important", mcp.Description("Mark matching messages as important")),
        mcp.WithBoolean("mark_read", mcp.Description("Mark matching messages as read")),
        mcp.WithBoolean("archive", mcp.Description("Archive matching messages")),
    )
    s.AddTool(createFilterTool, util.ErrorGuard(gmailCreateFilterHandler))

    // List filters tool
    listFiltersTool := mcp.NewTool("gmail_list_filters",
        mcp.WithDescription("List all Gmail filters in the account"),
    )
    s.AddTool(listFiltersTool, util.ErrorGuard(gmailListFiltersHandler))

    // List labels tool
    listLabelsTool := mcp.NewTool("gmail_list_labels",
        mcp.WithDescription("List all Gmail labels in the account"),
    )
    s.AddTool(listLabelsTool, util.ErrorGuard(gmailListLabelsHandler))

    // Add delete filter tool
    deleteFilterTool := mcp.NewTool("gmail_delete_filter",
        mcp.WithDescription("Delete a Gmail filter by its ID"),
        mcp.WithString("filter_id", mcp.Required(), mcp.Description("The ID of the filter to delete")),
    )
    s.AddTool(deleteFilterTool, util.ErrorGuard(gmailDeleteFilterHandler))

    // Add delete label tool
    deleteLabelTool := mcp.NewTool("gmail_delete_label",
        mcp.WithDescription("Delete a Gmail label by its ID"),
        mcp.WithString("label_id", mcp.Required(), mcp.Description("The ID of the label to delete")),
    )
    s.AddTool(deleteLabelTool, util.ErrorGuard(gmailDeleteLabelHandler))
}

var gmailService = sync.OnceValue[*gmail.Service](func() *gmail.Service {
	ctx := context.Background()

    tokenFile := os.Getenv("GOOGLE_TOKEN_FILE")
	if tokenFile == "" {
		panic("GOOGLE_TOKEN_FILE environment variable must be set")
	}

	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		panic("GOOGLE_CREDENTIALS_FILE environment variable must be set")
	}

	client := services.GoogleHttpClient(tokenFile, credentialsFile)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic(fmt.Sprintf("failed to create Gmail service: %v", err))
	}

	return srv
})

func gmailSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    query, ok := arguments["query"].(string)
    if !ok {
        return mcp.NewToolResultError("query must be a string"), nil
    }

    user := "me"
    
    listCall := gmailService().Users.Messages.List(user).Q(query).MaxResults(10)
    
    resp, err := listCall.Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to search emails: %v", err)), nil
    }

    var result strings.Builder
    result.WriteString(fmt.Sprintf("Found %d emails:\n\n", len(resp.Messages)))

    for _, msg := range resp.Messages {
        message, err := gmailService().Users.Messages.Get(user, msg.Id).Do()
        if err != nil {
            log.Printf("Failed to get message %s: %v", msg.Id, err)
            continue
        }

        details := make(map[string]string)
        for _, header := range message.Payload.Headers {
            switch header.Name {
            case "From":
                details["from"] = header.Value
            case "Subject":
                details["subject"] = header.Value
            case "Date":
                details["date"] = header.Value
            }
        }

        result.WriteString(fmt.Sprintf("Message ID: %s\n", msg.Id))
        result.WriteString(fmt.Sprintf("From: %s\n", details["from"]))
        result.WriteString(fmt.Sprintf("Subject: %s\n", details["subject"]))
        result.WriteString(fmt.Sprintf("Date: %s\n", details["date"]))
        result.WriteString(fmt.Sprintf("Snippet: %s\n", message.Snippet))
        result.WriteString("-------------------\n")
    }

    return mcp.NewToolResultText(result.String()), nil
}

func gmailMoveToSpamHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    messageIdsStr, ok := arguments["message_ids"].(string)
    if !ok {
        return mcp.NewToolResultError("message_ids must be a string"), nil
    }

    messageIds := strings.Split(messageIdsStr, ",")

    if len(messageIds) == 0 {
        return mcp.NewToolResultError("no message IDs provided"), nil
    }

    user := "me"

    for _, messageId := range messageIds {
        _, err := gmailService().Users.Messages.Modify(user, messageId, &gmail.ModifyMessageRequest{
            AddLabelIds: []string{"SPAM"},
        }).Do()
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("failed to move email %s to spam: %v", messageId, err)), nil
        }
    }

    return mcp.NewToolResultText(fmt.Sprintf("Successfully moved %d emails to spam.", len(messageIds))), nil
}

func gmailCreateFilterHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    // Create filter criteria
    criteria := &gmail.FilterCriteria{}
    
    if from, ok := arguments["from"].(string); ok && from != "" {
        criteria.From = from
    }
    if to, ok := arguments["to"].(string); ok && to != "" {
        criteria.To = to
    }
    if subject, ok := arguments["subject"].(string); ok && subject != "" {
        criteria.Subject = subject
    }
    if query, ok := arguments["query"].(string); ok && query != "" {
        criteria.Query = query
    }

    // Create filter action
    action := &gmail.FilterAction{}

    if addLabel, ok := arguments["add_label"].(bool); ok && addLabel {
        labelName, ok := arguments["label_name"].(string)
        if !ok || labelName == "" {
            return mcp.NewToolResultError("label_name is required when add_label is true"), nil
        }

        // First, create or get the label
        label, err := createOrGetLabel(labelName)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("failed to create/get label: %v", err)), nil
        }
        action.AddLabelIds = []string{label.Id}
    }

    if markImportant, ok := arguments["mark_important"].(bool); ok && markImportant {
        action.AddLabelIds = append(action.AddLabelIds, "IMPORTANT")
    }

    if markRead, ok := arguments["mark_read"].(bool); ok && markRead {
        action.RemoveLabelIds = append(action.RemoveLabelIds, "UNREAD")
    }

    if archive, ok := arguments["archive"].(bool); ok && archive {
        action.RemoveLabelIds = append(action.RemoveLabelIds, "INBOX")
    }

    // Create the filter
    filter := &gmail.Filter{
        Criteria: criteria,
        Action:   action,
    }

    result, err := gmailService().Users.Settings.Filters.Create("me", filter).Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to create filter: %v", err)), nil
    }

    return mcp.NewToolResultText(fmt.Sprintf("Successfully created filter with ID: %s", result.Id)), nil
}

func createOrGetLabel(name string) (*gmail.Label, error) {
    // First try to find existing label
    labels, err := gmailService().Users.Labels.List("me").Do()
    if err != nil {
        return nil, fmt.Errorf("failed to list labels: %v", err)
    }

    for _, label := range labels.Labels {
        if label.Name == name {
            return label, nil
        }
    }

    // If not found, create new label
    newLabel := &gmail.Label{
        Name:                  name,
        MessageListVisibility: "show",
        LabelListVisibility:   "labelShow",
    }

    label, err := gmailService().Users.Labels.Create("me", newLabel).Do()
    if err != nil {
        return nil, fmt.Errorf("failed to create label: %v", err)
    }

    return label, nil
}

func gmailListFiltersHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    filters, err := gmailService().Users.Settings.Filters.List("me").Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to list filters: %v", err)), nil
    }

    var result strings.Builder
    result.WriteString(fmt.Sprintf("Found %d filters:\n\n", len(filters.Filter)))

    for _, filter := range filters.Filter {
        result.WriteString(fmt.Sprintf("Filter ID: %s\n", filter.Id))
        
        // Write criteria
        result.WriteString("Criteria:\n")
        if filter.Criteria.From != "" {
            result.WriteString(fmt.Sprintf("  From: %s\n", filter.Criteria.From))
        }
        if filter.Criteria.To != "" {
            result.WriteString(fmt.Sprintf("  To: %s\n", filter.Criteria.To))
        }
        if filter.Criteria.Subject != "" {
            result.WriteString(fmt.Sprintf("  Subject: %s\n", filter.Criteria.Subject))
        }
        if filter.Criteria.Query != "" {
            result.WriteString(fmt.Sprintf("  Query: %s\n", filter.Criteria.Query))
        }

        // Write actions
        result.WriteString("Actions:\n")
        if len(filter.Action.AddLabelIds) > 0 {
            result.WriteString(fmt.Sprintf("  Add Labels: %s\n", strings.Join(filter.Action.AddLabelIds, ", ")))
        }
        if len(filter.Action.RemoveLabelIds) > 0 {
            result.WriteString(fmt.Sprintf("  Remove Labels: %s\n", strings.Join(filter.Action.RemoveLabelIds, ", ")))
        }
        result.WriteString("-------------------\n")
    }

    return mcp.NewToolResultText(result.String()), nil
}

func gmailListLabelsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    labels, err := gmailService().Users.Labels.List("me").Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to list labels: %v", err)), nil
    }

    var result strings.Builder
    result.WriteString(fmt.Sprintf("Found %d labels:\n\n", len(labels.Labels)))

    // First list system labels
    result.WriteString("System Labels:\n")
    for _, label := range labels.Labels {
        if label.Type == "system" {
            result.WriteString(fmt.Sprintf("- %s (ID: %s)\n", label.Name, label.Id))
        }
    }

    // Then list user labels
    result.WriteString("\nUser Labels:\n")
    for _, label := range labels.Labels {
        if label.Type == "user" {
            result.WriteString(fmt.Sprintf("- %s (ID: %s)\n", label.Name, label.Id))
            if label.MessagesTotal > 0 {
                result.WriteString(fmt.Sprintf("  Messages: %d\n", label.MessagesTotal))
            }
        }
    }

    return mcp.NewToolResultText(result.String()), nil
}

func gmailDeleteFilterHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    filterID, ok := arguments["filter_id"].(string)
    if !ok {
        return mcp.NewToolResultError("filter_id must be a string"), nil
    }

    if filterID == "" {
        return mcp.NewToolResultError("filter_id cannot be empty"), nil
    }

    err := gmailService().Users.Settings.Filters.Delete("me", filterID).Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to delete filter: %v", err)), nil
    }

    return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted filter with ID: %s", filterID)), nil
}

func gmailDeleteLabelHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	labelID, ok := arguments["label_id"].(string)
	if !ok {
		return mcp.NewToolResultError("label_id must be a string"), nil
	}

	if labelID == "" {
		return mcp.NewToolResultError("label_id cannot be empty"), nil
	}

	err := gmailService().Users.Labels.Delete("me", labelID).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete label: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted label with ID: %s", labelID)), nil
}

func gmailReadEmailHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    messageID, ok := arguments["message_id"].(string)
    if !ok {
        return mcp.NewToolResultError("message_id must be a string"), nil
    }

    includeAttachments, _ := arguments["include_attachments"].(bool)

    // Get the full email message
    message, err := gmailService().Users.Messages.Get("me", messageID).Format("full").Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to get email: %v", err)), nil
    }

    var result strings.Builder

    // Extract headers
    headers := make(map[string]string)
    for _, header := range message.Payload.Headers {
        switch header.Name {
        case "From", "To", "Cc", "Subject", "Date":
            headers[header.Name] = header.Value
            result.WriteString(fmt.Sprintf("%s: %s\n", header.Name, header.Value))
        }
    }
    result.WriteString("\n")

    // Extract body
    body := extractMessageBody(message.Payload)
    result.WriteString("Body:\n")
    result.WriteString(body)
    result.WriteString("\n")

    // Handle attachments if requested
    if includeAttachments && len(message.Payload.Parts) > 0 {
        result.WriteString("\nAttachments:\n")
        for _, part := range message.Payload.Parts {
            if part.Filename != "" {
                result.WriteString(fmt.Sprintf("- %s (Size: %d bytes)\n", 
                    part.Filename, part.Body.Size))
            }
        }
    }

    return mcp.NewToolResultText(result.String()), nil
}

func extractMessageBody(payload *gmail.MessagePart) string {
    if payload.MimeType == "text/plain" && payload.Body.Data != "" {
        data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
        if err != nil {
            return fmt.Sprintf("Error decoding body: %v", err)
        }
        return string(data)
    }

    if payload.Parts != nil {
        for _, part := range payload.Parts {
            if part.MimeType == "text/plain" {
                data, err := base64.URLEncoding.DecodeString(part.Body.Data)
                if err != nil {
                    continue
                }
                return string(data)
            }
        }
    }

    return "No readable text body found"
}

func gmailReplyEmailHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    messageID, ok := arguments["message_id"].(string)
    if !ok {
        return mcp.NewToolResultError("message_id must be a string"), nil
    }

    replyText, ok := arguments["reply_text"].(string)
    if !ok {
        return mcp.NewToolResultError("reply_text must be a string"), nil
    }

    replyAll, _ := arguments["reply_all"].(bool)

    // Get the original message to extract headers
    originalMessage, err := gmailService().Users.Messages.Get("me", messageID).Format("metadata").Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to get original email: %v", err)), nil
    }

    // Extract necessary headers
    var from, to, subject, references, messageIDHeader string
    for _, header := range originalMessage.Payload.Headers {
        switch header.Name {
        case "From":
            to = header.Value // Original sender becomes recipient
        case "To":
            from = header.Value // We'll need this for reply-all
        case "Subject":
            subject = header.Value
            if !strings.HasPrefix(strings.ToLower(subject), "re:") {
                subject = "Re: " + subject
            }
        case "Message-ID":
            messageIDHeader = header.Value
            references = header.Value
        case "References":
            references = header.Value + " " + messageIDHeader
        }
    }

    // Create reply message
    var message gmail.Message

    // Prepare recipients
    recipients := []string{to}
    if replyAll {
        // Add original To recipients (excluding ourselves)
        originalRecipients := strings.Split(from, ",")
        for _, recipient := range originalRecipients {
            recipient = strings.TrimSpace(recipient)
            if recipient != "" && !strings.Contains(recipient, "me@") {
                recipients = append(recipients, recipient)
            }
        }
    }

    // Construct email headers
    headers := make(map[string]string)
    headers["To"] = strings.Join(recipients, ", ")
    headers["Subject"] = subject
    headers["References"] = references
    headers["In-Reply-To"] = messageIDHeader

    // Construct the raw message
    var rawMessage strings.Builder
    for key, value := range headers {
        rawMessage.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
    }
    rawMessage.WriteString("\r\n")
    rawMessage.WriteString(replyText)

    // Encode the raw message
    message.Raw = base64.URLEncoding.EncodeToString([]byte(rawMessage.String()))

    // Send the reply
    _, err = gmailService().Users.Messages.Send("me", &message).Do()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to send reply: %v", err)), nil
    }

    return mcp.NewToolResultText("Reply sent successfully"), nil
}