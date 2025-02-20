package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type EnvVarInfo struct {
    comment string
}

type ToolInfo struct {
    Name        string
    Description string
    FileName    string
}

func updateReadmeConfig(envVars map[string]EnvVarInfo, tools []ToolInfo) error {
    // Read README.md
    content, err := ioutil.ReadFile("README.md")
    if err != nil {
        return fmt.Errorf("error reading README.md: %v", err)
    }

    // Convert to string
    readmeContent := string(content)

    // Update env vars section
    configRegex := regexp.MustCompile(`(?s)"env": \{[^}]*\}`)
    
    var envConfig strings.Builder
    envConfig.WriteString(`"env": {`)
    first := true
    for envVar, info := range envVars {
        if !first {
            envConfig.WriteString(",")
        }
        first = false
        envConfig.WriteString("\n        ")
        envConfig.WriteString(fmt.Sprintf(`"%s": "%s"`, envVar, info.comment))
    }
    envConfig.WriteString("\n      }")

    // Replace env config
    readmeContent = configRegex.ReplaceAllString(readmeContent, envConfig.String())

    // Group tools by filename
    toolsByFile := make(map[string][]ToolInfo)
    for _, tool := range tools {
        toolsByFile[tool.FileName] = append(toolsByFile[tool.FileName], tool)
    }

    // Generate tools section content
    var toolsSection strings.Builder
    toolsSection.WriteString("## Available Tools\n\n")
    
    // Sort filenames for consistent output
    var fileNames []string
    for fileName := range toolsByFile {
        fileNames = append(fileNames, fileName)
    }
    sort.Strings(fileNames)

    for _, fileName := range fileNames {
        toolsSection.WriteString(fmt.Sprintf("### Group: %s\n\n", strings.TrimSuffix(fileName, ".go")))
        for _, tool := range toolsByFile[fileName] {
            toolsSection.WriteString(fmt.Sprintf("#### %s\n\n", tool.Name))
            if tool.Description != "" {
                toolsSection.WriteString(fmt.Sprintf("%s\n\n", tool.Description))
            }
        }
    }

    // Replace existing tools section
    // Look for the section between "## Available Tools" and the next section starting with "##"
    toolsSectionRegex := regexp.MustCompile(`(?s)## Available Tools.*?(\n## |$)`)
    
    if toolsSectionRegex.MatchString(readmeContent) {
        // Replace existing section
        readmeContent = toolsSectionRegex.ReplaceAllString(readmeContent, toolsSection.String())
    } else {
        // If section doesn't exist, add it before the end
        readmeContent += "\n\n" + toolsSection.String()
    }

    // Write back to README.md
    err = ioutil.WriteFile("README.md", []byte(readmeContent), 0644)
    if err != nil {
        return fmt.Errorf("error writing README.md: %v", err)
    }

    return nil
}

func extractToolInfo(node *ast.File, fileName string) []ToolInfo {
    var tools []ToolInfo

    ast.Inspect(node, func(n ast.Node) bool {
        // Look for tool registrations
        callExpr, ok := n.(*ast.CallExpr)
        if !ok {
            return true
        }

        // Check if it's a NewTool call
        if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
            if sel.Sel.Name == "NewTool" {
                tool := ToolInfo{
                    FileName: fileName,
                }
                
                // Extract tool name
                if len(callExpr.Args) > 0 {
                    if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
                        tool.Name = strings.Trim(lit.Value, `"'`)
                    }
                }

                // Extract description and arguments from WithX calls
                for _, arg := range callExpr.Args[1:] {
                    if call, ok := arg.(*ast.CallExpr); ok {
                        if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                            switch sel.Sel.Name {
                            case "WithDescription":
                                if len(call.Args) > 0 {
                                    if lit, ok := call.Args[0].(*ast.BasicLit); ok {
                                        tool.Description = strings.Trim(lit.Value, `"'`)
                                    }
                                }
                            }
                        }
                    }
                }

                if tool.Name != "" {
                    tools = append(tools, tool)
                }
            }
        }
        return true
    })

    return tools
}

func main() {
    envVars := make(map[string]EnvVarInfo)
    var allTools []ToolInfo

    // Walk through all .go files
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !strings.HasSuffix(path, ".go") {
            return nil
        }

        // Parse the Go file
        fset := token.NewFileSet()
        node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
        if err != nil {
            return fmt.Errorf("error parsing %s: %v", path, err)
        }

        // Extract environment variables
        ast.Inspect(node, func(n ast.Node) bool {
            call, ok := n.(*ast.CallExpr)
            if !ok {
                return true
            }

            sel, ok := call.Fun.(*ast.SelectorExpr)
            if !ok {
                return true
            }

            if ident, ok := sel.X.(*ast.Ident); ok {
                if ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv") {
                    if len(call.Args) > 0 {
                        if strLit, ok := call.Args[0].(*ast.BasicLit); ok && strLit.Kind == token.STRING {
                            envName := strings.Trim(strLit.Value, `"'`)
                            
                            var comment string
                            for _, cg := range node.Comments {
                                if cg.End() < call.Pos() {
                                    lastComment := cg.List[len(cg.List)-1]
                                    if lastComment.End()+100 >= call.Pos() {
                                        comment = strings.TrimPrefix(lastComment.Text, "//")
                                        comment = strings.TrimSpace(comment)
                                    }
                                }
                            }

                            envVars[envName] = EnvVarInfo{comment: comment}
                        }
                    }
                }
            }
            return true
        })

        // Extract tool information if in tools directory
        if strings.HasPrefix(path, "tools/") {
            fileName := filepath.Base(path)
            tools := extractToolInfo(node, fileName)
            allTools = append(allTools, tools...)
        }

        return nil
    })

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error walking files: %v\n", err)
        os.Exit(1)
    }

    // Update README.md with both env vars and tools
    err = updateReadmeConfig(envVars, allTools)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error updating README.md: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Successfully updated README.md with environment variables and tools documentation")
}