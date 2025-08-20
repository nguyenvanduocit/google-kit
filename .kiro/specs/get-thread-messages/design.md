# Design

Implement a new `gchat_get_thread_messages` tool that follows the existing pattern in `gchat_list_messages`.

## Implementation

1. **Tool Registration**: Add to `RegisterGChatTool()` function
   - Parameters: space_name (required), thread_name (required), page_size (optional, default 100), page_token (optional)
   - Handler: `gChatGetThreadMessagesHandler` wrapped with `util.ErrorGuard`

2. **Handler Function**: Similar to `gChatListMessagesHandler` but with thread filtering
   - Extract parameters using type assertions
   - Use Google Chat API with thread filter: `Filter(fmt.Sprintf("thread.name = %s", threadName))`
   - Return YAML response with same message structure as existing tools

3. **Response Format**: Match existing gchat tools
   ```yaml
   messages:
     - name: "..."
       sender: {...}
       createTime: "..."
       text: "..."
       thread: {...}
   nextPageToken: "..."
   ```

## Code Structure
Follow the exact same pattern as `gChatListMessagesHandler` - simple, straightforward implementation without over-engineering.