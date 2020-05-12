set -eo pipefail

USER_ID="$1"
GROUP_ID="$2"

# Build public frontend
echo "Build public frontend..."
nix-build -A frontend-public

mkdir -p build/frontend_public
cp result ./build/frontend_public/frontend_public.js
rm result
echo "Done building public frontend"

# Build backend
echo "Build  backend..."
mkdir -p backend/vendor

nix-shell --run 'cd backend && go mod vendor'
nix-build -A backend-image

mkdir -p build/backend
cp result ./build/backend/backend-image.tar.gz
rm result
echo "Done building backend"

chown -R "$USER_ID:$GROUP_ID" build