# Public export
export DB_USER=kevin
export DEFAULT_DB_HOST=localhost
export DEFAULT_DB_PORT=5432

# Private export
if [ -f "$HOME/.postgresql" ]; then
  source "$HOME/.postgresql"
else
  warn "File ~/.postgresql required to instanciate password."
fi
