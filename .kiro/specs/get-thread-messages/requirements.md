# Requirements

Add a new MCP tool `gchat_get_thread_messages` to retrieve messages from specific Google Chat threads.

## Core Requirements

1. **Thread Message Retrieval**: Get messages from a specific thread using space_name and thread_name parameters
2. **Pagination**: Support page_size and page_token for large threads (default 100 messages)
3. **Response Format**: Return YAML format matching existing gchat tools with message details and attachments
4. **Error Handling**: Clear error messages for missing/invalid parameters and API failures
5. **Integration**: Follow existing patterns in gchat.go with ErrorGuard wrapper