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
    export PATH=$GOPATH/bin:$PWD/scripts:$PATH
    
    echo ""
    echo "🏛️  Traditional Builders - Dev Environment"
    echo "=========================================="
    echo ""
    echo "Environment:"
    echo "  Go      $(go version | awk '{print $3}')"
    echo "  Goose   $(goose -version 2>&1 | head -n1 | awk '{print $NF}')"
    echo "  Templ   $(templ version 2>&1 | head -n1 || echo 'installed')"
    echo ""
    echo "Commands:"
    echo "  setup          - Initialize project"
    echo "  run            - Start development server"
    echo "  build          - Build production binary"
    echo "  test           - Run tests"
    echo "  migrate        - Run database migrations"
    echo "  migrate-down   - Rollback last migration"
    echo "  clean          - Remove generated files"
    echo ""
  '';
}
