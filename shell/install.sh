#!/bin/bash

# Build the binary
go build -o cogito ./cmd/cogito

# Create config directory
mkdir -p ~/.cogito

# Copy default config
cp configs/default.yaml ~/.cogito/config.yaml

# Create default context file
cat > ~/.cogito/context.md << 'EOF'
# Cogito Context

## Project Info
- Working Directory: {{cwd}}
- Session: {{session_id}}

## Notes
Add your project-specific context here.

## Memory
Memory features coming soon.
EOF

echo "✅ Cogito installed successfully!"
echo "📍 Binary: $(pwd)/cogito"
echo "📁 Config: ~/.cogito/config.yaml"
echo ""
echo "Usage with Claude Code:"
echo "  claude hook claude-code session_start < cogito --platform claude-code --event session_start"
