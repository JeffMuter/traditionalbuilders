{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    goose
    sqlite
    templ
    nodejs
    nodePackages.tailwindcss
    gnumake
  ];

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PATH
    
    echo ""
    echo "🏛️  Traditional Builders - Dev Environment"
    echo "=========================================="
    echo ""
    echo "Environment:"
    echo "  Go      $(go version | awk '{print $3}')"
    echo "  Goose   $(goose -version 2>&1 | head -n1 | awk '{print $NF}')"
    echo "  Templ   $(templ version 2>&1 | head -n1 || echo 'installed')"
    echo ""
    echo "Quick Start:"
    echo "  make setup   - Initialize project"
    echo "  make dev     - Start development server"
    echo "  make help    - Show all commands"
    echo ""
  '';
}
