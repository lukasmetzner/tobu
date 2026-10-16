unset ZELLIJ ZELLIJ_SESSION_NAME ZELLIJ_PANE_ID TMUX

export GOCACHE=$(go env GOCACHE)
export DEMO=/tmp/tobu-demo
export _ZO_DATA_DIR=$DEMO/zoxide
export ZELLIJ_SOCKET_DIR=$DEMO/zellij
export XDG_CACHE_HOME=$DEMO/cache
export TOBU_CONFIG=$DEMO/config.toml
export ZELLIJ_CONFIG_FILE=$DEMO/zellij.kdl
export PATH=$DEMO/bin:$PATH

[ -d "$DEMO" ] && rm -rf "$DEMO"
mkdir -p "$DEMO/bin" "$_ZO_DATA_DIR" "$ZELLIJ_SOCKET_DIR"
mkdir -p "$XDG_CACHE_HOME/zellij/$(zellij --version | cut -d" " -f2)"
touch "$XDG_CACHE_HOME/zellij/$(zellij --version | cut -d" " -f2)/seen_release_notes"
go build -o "$DEMO/bin/tobu" .

{
  grep -v '^theme ' "${XDG_CONFIG_HOME:-$HOME/.config}/zellij/config.kdl"
  echo 'theme "tokyo-night-dark"'
} > "$ZELLIJ_CONFIG_FILE"

cat > "$TOBU_CONFIG" <<'TOML'
[editor]
binary = "nvim"
args = ["-c", "colorscheme tokyonight-night"]
TOML

visit() {
  mkdir -p "$DEMO/$1"
  touch "$DEMO/$1/README.md"
  for _ in $(seq "$2"); do zoxide add "$DEMO/$1"; done
}

visit src/github.com/acme/api 12
visit src/github.com/acme/web 9
visit src/github.com/acme/infra 7
visit src/github.com/acme/cli 5
visit src/github.com/acme/docs 3
visit dotfiles 4
visit notes 2

(cd "$DEMO/src/github.com/acme/api" && zellij attach --create-background api)

cd "$DEMO"
PS1='$ '
clear
