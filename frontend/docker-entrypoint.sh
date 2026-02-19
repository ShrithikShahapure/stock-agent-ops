#!/bin/sh
# Write runtime API_URL into config.js so the SPA picks it up
API_URL="${API_URL:-http://localhost:8000}"
cat > /usr/share/nginx/html/config.js <<EOF
window.__API_URL__ = "${API_URL}";
EOF

exec "$@"
