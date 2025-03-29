# Coder

Trying to build a basic code analysis CLI (to learn Go along the way). 

## Getting Started

### Prerequisites

- Go (> 1.18)
- OpenAI API key

### Installation

1. Clone the repository:
   ```bash
   gh repo clone varunrandery/coder
   ```

2. Navigate to the project directory:
   ```bash
   cd coder
   ```

3. Write your OpenAI API key to .env:
   ```bash
   echo "OPENAI_API_KEY=<key>" > .env
   ```

4. Run with `go run .`!

## Usage

Output of `/help`:
```
Usage:
- Type your message and press Enter to get a response.
- "/new": start a new conversation.
- "/include <file-path> <prompt>": include a file in context.
- "/session": view current-conversation token consumption.
- "/model info": to view current model info.
- "/model switch <model-name>" to change the current model.
- "/exit": exit the program.
```

## License

MIT License.