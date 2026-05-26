{ pkgs ? import <nixpkgs> {
    config = {
      allowUnfree = true;
      allowUnsupportedSystem = true;
    };
  }
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    goose
    sqlite
    templ
    nodejs
    nodePackages.tailwindcss
    gnumake
    air       # hot-reload: rebuilds + restarts server on .go changes
    overmind  # process manager: runs templ/css/air together from Procfile
    tmux      # required by overmind
  ];

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PWD/scripts:$PATH

    # Set writable npm global prefix and add to PATH
    export NPM_PREFIX="$HOME/.npm-global"
    mkdir -p "$NPM_PREFIX/bin"
    npm config set prefix "$NPM_PREFIX"
    export PATH="$NPM_PREFIX/bin:$PATH"

    # Install claude-code via npm (avoids nixpkgs version lag)
    if ! command -v claude &> /dev/null; then
      echo "Installing claude-code via npm..."
      npm install -g @anthropic-ai/claude-code --silent
    fi

    echo ""
    echo "🏛️  Traditional Builders - Dev Environment"
    echo "=========================================="
    echo ""
    echo "Environment:"
    echo "  Go       $(go version | awk '{print $3}')"
    echo "  Goose    $(goose -version 2>&1 | head -n1 | awk '{print $NF}')"
    echo "  Templ    $(templ version 2>&1 | head -n1 || echo 'installed')"
    echo "  Air      $(air -v 2>&1 | head -n1 || echo 'installed')"
    echo "  Overmind $(overmind --version 2>/dev/null || echo 'installed')"
    echo "  Claude   $(claude --version 2>/dev/null || echo 'not found')"
    echo ""
    echo "Commands:"
    echo "  dev-watch        - Live-reload dev server (recommended)"
    echo "  run              - One-shot dev server (no live reload)"
    echo "  setup            - Initialize project (migrate + css + templ)"
    echo "  build            - Build production binary"
    echo "  test             - Run tests"
    echo "  migrate          - Run database migrations"
    echo "  migrate-down     - Rollback last migration"
    echo "  reset-db         - Drop and re-run all migrations"
    echo "  seed-zip-codes   - Import full GeoNames US zip code dataset"
    echo "  clean            - Remove generated files"
    echo ""
  '';
}
