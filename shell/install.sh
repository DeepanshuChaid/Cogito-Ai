#!/bin/bash

echo "🦍 Installing Cogito..."

# Build binary
go build -o cogito ./cmd/cogito

# Move to PATH
sudo mv cogito /usr/local/bin/

# Create config directory
mkdir -p ~/.cogito

# Create default context
cat > ~/.cogito/context.md << 'EOF'
# Cogito Context

## Instructions
- Be terse and direct
- Drop filler words
- Technical accuracy > politeness

## Project
Add your project-specific context here.
EOF

# Copy hook file
HOOKS_DIR="$PWD/.codex"
mkdir -p "$HOOKS_DIR"
cp hooks/codex-hooks.json "$HOOKS_DIR/hooks.json"

echo "✅ Cogito installed!"
echo "📍 Binary: /usr/local/bin/cogito"
echo "📁 Config: ~/.cogito/context.md"
echo "🔗 Hook: $HOOKS_DIR/hooks.json"
echo ""
echo "Now run Codex in this directory:"
echo "  codex"
