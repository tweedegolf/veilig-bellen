USER_ID="$1"
GROUP_ID="$2"

# Build public frontend
nix-build -A frontend-public

mkdir -p build/frontend_public
cp result ./build/frontend_public/frontend_public.js
rm result

# Build backend
nix-build -A backend-image

mkdir -p build/backend
cp result ./build/backend/backend-image.tar.gz
rm result

chown -R "$USER_ID:$GROUP_ID" build