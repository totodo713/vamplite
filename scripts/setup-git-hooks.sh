#!/bin/bash

# Git hooks setup script for Muscle Dreamer project
# This script sets up pre-commit hooks and git configuration

set -e

echo "🔧 Setting up Git hooks and configuration for Muscle Dreamer..."

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# 1. Setup commit message template
echo "📝 Setting up commit message template..."
git config commit.template .gitmessage
echo "✅ Commit template configured"

# 2. Create pre-commit hook
echo "🔍 Setting up pre-commit hook..."
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Pre-commit hook for Muscle Dreamer project
# Runs format, lint, and test before allowing commit

set -e

echo "🔍 Running pre-commit checks..."

# Check if make commands are available
if ! command -v make &> /dev/null; then
    echo "❌ Error: make command not found"
    exit 1
fi

# 1. Format code
echo "📝 Formatting code..."
if ! make format; then
    echo "❌ Code formatting failed"
    exit 1
fi

# 2. Run linter
echo "🔍 Running linter..."
if ! make lint; then
    echo "❌ Linting failed"
    echo "Please fix all warnings and errors before committing"
    exit 1
fi

# 3. Run tests
echo "🧪 Running tests..."
if ! make test; then
    echo "❌ Tests failed"
    echo "Please fix all failing tests before committing"
    exit 1
fi

echo "✅ All pre-commit checks passed!"
echo "🚀 Ready to commit"
EOF

# Make the hook executable
chmod +x .git/hooks/pre-commit
echo "✅ Pre-commit hook installed"

# 3. Configure branch naming convention reminder
echo "🌿 Setting up branch naming convention..."
git config alias.new-feature '!f() { git checkout -b "developer/$1"; }; f'
echo "✅ Branch alias configured (use: git new-feature feature-name)"

# 4. Setup useful aliases
echo "🔧 Setting up useful Git aliases..."
git config alias.st status
git config alias.co checkout
git config alias.br branch
git config alias.ci commit
git config alias.unstage 'reset HEAD --'
git config alias.last 'log -1 HEAD'
git config alias.visual '!gitk'
echo "✅ Git aliases configured"

echo ""
echo "🎉 Git setup completed successfully!"
echo ""
echo "📋 What was configured:"
echo "   ✅ Commit message template (.gitmessage)"
echo "   ✅ Pre-commit hook (format + lint + test)"
echo "   ✅ Branch naming alias (git new-feature <name>)"
echo "   ✅ Useful Git aliases"
echo ""
echo "📖 Usage examples:"
echo "   git new-feature my-awesome-feature  # Creates developer/my-awesome-feature branch"
echo "   git commit                          # Uses Japanese template"
echo "   git st                              # Short for git status"
echo ""
echo "⚠️  Remember: Pre-commit hook will run 'make format && make lint && make test'"
echo "    Make sure these commands work in your development environment!"