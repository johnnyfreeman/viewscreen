Run `claude -p ... --output-format stream-json --verbose` and document all the different kinds of output objects you encounter. Your goal is to create an accurate spec documents that lists all types of output and each of their properties as tersely as possible. If documents don't exist yet, create them. Each output block type should have it's own file with documentation/spec.

IMPORTANT: Use subagents to do the actual command execution and data gathering. You should:
1. Spawn a subagent to run various claude commands and capture the JSON output types
2. Have the subagent write discovered types/properties directly to files in docs/stream-json/
3. Only receive summaries back, not full JSON payloads
4. This prevents filling up the main agent's context with verbose JSON output 
