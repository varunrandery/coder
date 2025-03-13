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

Basic slash commands are supported; attach files using `\attach <file-path>` to include the file in the next prompt. Use `\new` to start a new conversation and `\exit` to cleanly exit.

## License

MIT License.