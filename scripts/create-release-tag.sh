#!/bin/bash
# create-release-tag.sh
# Creates a semantic version tag for release
# Usage: ./create-release-tag.sh [patch|minor|major]

set -e

BUMP_TYPE="${1:-patch}"

# Validate input
case "$BUMP_TYPE" in
    patch|minor|major)
        ;;
    *)
        echo "Usage: $0 [patch|minor|major]"
        exit 1
        ;;
esac

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Remove 'v' prefix for processing
VERSION="${LATEST_TAG#v}"

# Split version into parts
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"
MAJOR=${MAJOR:-0}
MINOR=${MINOR:-0}
PATCH=${PATCH:-0}

# Bump version based on type
case "$BUMP_TYPE" in
    patch)
        PATCH=$((PATCH + 1))
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
esac

NEW_TAG="v${MAJOR}.${MINOR}.${PATCH}"

echo "Bumping version from $LATEST_TAG to $NEW_TAG"

# Create and push tag
git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
echo "Created tag: $NEW_TAG"

git push origin "$NEW_TAG"
echo "Pushed tag to remote"

echo "New version: $NEW_TAG"
