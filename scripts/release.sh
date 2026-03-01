#!/bin/bash
set -euo pipefail

# ─── Nest CLI: Build & Release Script ───────────────────────────
# Reads version from git tag, cross-compiles for all platforms,
# and creates a GitHub release with binary artifacts.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
RELEASE_REPO="koyeo/nest"
APP_NAME="nest"

# 1. Determine version
if [[ "${1:-}" != "" ]]; then
    VERSION="$1"
else
    VERSION=$(git -C "$ROOT_DIR" describe --tags --abbrev=0 2>/dev/null || echo "")
    if [[ -z "$VERSION" ]]; then
        echo "❌ No version specified and no git tag found."
        echo "   Usage: $0 <version>   (e.g. $0 v0.1.0)"
        exit 1
    fi
fi
# Ensure version starts with 'v'
[[ "$VERSION" != v* ]] && VERSION="v$VERSION"
TAG="$VERSION"
COMMIT=$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "🚀 Nest Release: $TAG"
echo "────────────────────────────"

# 2. Ensure clean working tree
if ! git -C "$ROOT_DIR" diff --quiet || ! git -C "$ROOT_DIR" diff --cached --quiet; then
    echo "❌ Working tree is dirty. Please commit or stash your changes before releasing."
    exit 1
fi
echo "✅ Working tree is clean"

# 3. Check if local commits are pushed to remote
git -C "$ROOT_DIR" fetch --quiet
LOCAL_SHA=$(git -C "$ROOT_DIR" rev-parse HEAD)
REMOTE_SHA=$(git -C "$ROOT_DIR" rev-parse @{u} 2>/dev/null || echo "")
if [[ "$LOCAL_SHA" != "$REMOTE_SHA" ]]; then
    echo "❌ Local branch is not in sync with remote. Please push your commits first."
    echo "   Local:  $LOCAL_SHA"
    echo "   Remote: $REMOTE_SHA"
    exit 1
fi
echo "✅ Local commits pushed to remote"

# 4. Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) not found. Install with: brew install gh"
    exit 1
fi

# 5. Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ GitHub CLI not authenticated. Run: gh auth login"
    exit 1
fi
echo "✅ GitHub CLI authenticated"

# 6. Check if tag already exists on the release repo
if gh release view "$TAG" --repo "$RELEASE_REPO" &> /dev/null; then
    echo "⚠️  Release $TAG already exists on $RELEASE_REPO. Delete it first or bump the version."
    echo "   To delete: gh release delete $TAG --repo $RELEASE_REPO --yes --cleanup-tag"
    exit 1
fi

# 7. Run tests
echo ""
echo "🧪 Running tests..."
(cd "$ROOT_DIR" && go test ./...)
echo "✅ Tests passed!"

# 8. Cross-compile for all platforms
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
LDFLAGS="-X main.version=$TAG -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME"

echo ""
echo "📦 Cross-compiling for all platforms..."
rm -rf "$BUILD_DIR/${APP_NAME}-"*
mkdir -p "$BUILD_DIR"

for platform in $PLATFORMS; do
    os="${platform%/*}"
    arch="${platform#*/}"
    output="${APP_NAME}-${os}-${arch}"
    if [[ "$os" == "windows" ]]; then
        output="${output}.exe"
    fi
    echo "   Building ${os}/${arch} → ${output}"
    GOOS="$os" GOARCH="$arch" go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/$output" "$ROOT_DIR" || exit 1
done
echo "✅ All binaries built!"

# 9. Generate checksums
echo ""
echo "🔐 Generating checksums..."
(cd "$BUILD_DIR" && shasum -a 256 ${APP_NAME}-* > checksums.txt)
cat "$BUILD_DIR/checksums.txt"

# 10. Collect artifacts
echo ""
echo "📋 Artifacts:"
ARTIFACTS=()
for f in "$BUILD_DIR"/${APP_NAME}-*; do
    if [[ -f "$f" ]]; then
        SIZE=$(du -h "$f" | cut -f1 | xargs)
        echo "   $(basename "$f")  ($SIZE)"
        ARTIFACTS+=("$f")
    fi
done
ARTIFACTS+=("$BUILD_DIR/checksums.txt")

if [[ ${#ARTIFACTS[@]} -eq 0 ]]; then
    echo "❌ No binary files found in $BUILD_DIR"
    exit 1
fi

# 11. Create git tag if it doesn't exist
if ! git -C "$ROOT_DIR" rev-parse "$TAG" &> /dev/null; then
    echo ""
    echo "🏷️  Creating git tag $TAG..."
    git -C "$ROOT_DIR" tag -a "$TAG" -m "Release $TAG"
    git -C "$ROOT_DIR" push origin "$TAG"
fi

# 12. Create GitHub release
echo ""
echo "🏷️  Creating GitHub release $TAG on $RELEASE_REPO..."
gh release create "$TAG" \
    "${ARTIFACTS[@]}" \
    --repo "$RELEASE_REPO" \
    --title "Nest $TAG" \
    --notes "## Nest $TAG

### Downloads

| Platform | Architecture | Binary |
|:---------|:-------------|:-------|
| macOS    | Apple Silicon (M1/M2/M3) | \`${APP_NAME}-darwin-arm64\` |
| macOS    | Intel        | \`${APP_NAME}-darwin-amd64\` |
| Linux    | x86_64       | \`${APP_NAME}-linux-amd64\` |
| Linux    | ARM64        | \`${APP_NAME}-linux-arm64\` |
| Windows  | x86_64       | \`${APP_NAME}-windows-amd64.exe\` |

### Quick Install

\`\`\`bash
# Using go install (requires Go)
go install github.com/koyeo/nest@$TAG

# Or download binary directly (macOS Apple Silicon example)
curl -fsSL https://github.com/$RELEASE_REPO/releases/download/$TAG/${APP_NAME}-darwin-arm64 -o /usr/local/bin/nest
chmod +x /usr/local/bin/nest
\`\`\`

### Checksums
See \`checksums.txt\` for SHA256 verification."

echo ""
echo "✅ Release $TAG published!"
echo "   https://github.com/$RELEASE_REPO/releases/tag/$TAG"
