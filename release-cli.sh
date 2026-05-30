#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Parse flags
BUMP_TYPE="patch"
VERSION=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --major)
            BUMP_TYPE="major"
            shift
            ;;
        --minor)
            BUMP_TYPE="minor"
            shift
            ;;
        --patch)
            BUMP_TYPE="patch"
            shift
            ;;
        v*)
            VERSION="$1"
            shift
            ;;
        *)
            error "Unknown argument: $1"
            ;;
    esac
done

# Auto-bump version if not provided
if [ -z "$VERSION" ]; then
    LATEST_TAG=$(git tag --sort=-v:refname | grep '^v[0-9]' | head -1)
    
    if [ -n "$LATEST_TAG" ]; then
        # Parse version (remove 'v' prefix)
        VERSION_NUM="${LATEST_TAG#v}"
        IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_NUM"
        
        case "$BUMP_TYPE" in
            major)
                MAJOR=$((MAJOR + 1))
                MINOR=0
                PATCH=0
                ;;
            minor)
                MINOR=$((MINOR + 1))
                PATCH=0
                ;;
            patch)
                PATCH=$((PATCH + 1))
                ;;
        esac
        
        VERSION="v${MAJOR}.${MINOR}.${PATCH}"
        info "Auto-bumped $LATEST_TAG → $VERSION ($BUMP_TYPE)"
    else
        VERSION="v0.1.0"
        info "No existing tags found, using initial version $VERSION"
    fi
    
    read -p "Proceed with version $VERSION? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Release aborted by user"
    fi
fi

# Validate version format (must start with 'v')
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format. Must be semver with 'v' prefix (e.g., v0.2.0)"
fi

info "Starting release for $VERSION"

# Check for uncommitted changes in cli/
if ! git diff --quiet cli/ || ! git diff --cached --quiet cli/; then
    warning "Uncommitted changes detected in cli/"
    read -p "Commit changes with message 'cli: changes for $VERSION'? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git add cli/
        git commit -m "cli: changes for $VERSION"
        success "Committed changes"
    else
        error "Release aborted. Please commit or stash changes first."
    fi
fi

# Ensure working directory is clean overall
if ! git diff --quiet || ! git diff --cached --quiet; then
    error "Working directory has uncommitted changes outside cli/. Please commit or stash them first."
fi

# Verify the CLI builds
info "Building CLI to verify..."
cd cli
if go build -o /dev/null ./cmd/fai; then
    success "CLI builds successfully"
else
    error "CLI build failed. Please fix build errors before releasing."
fi
cd ..

# Push to origin main
info "Pushing to origin/main..."
git push origin main
success "Pushed to origin/main"

# Create annotated tag
info "Creating annotated tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
success "Tag $VERSION created"

# Push the tag
info "Pushing tag $VERSION..."
git push origin "$VERSION"
success "Tag $VERSION pushed"

# Print summary
echo ""
success "Release $VERSION complete!"
echo ""
info "Summary:"
echo "  - Changes pushed to origin/main"
echo "  - Tag $VERSION created and pushed"
echo "  - GitHub Actions workflow should now trigger"
echo ""
info "Monitor the release at:"
echo "  https://github.com/ktappdev/fromai/actions"
echo ""