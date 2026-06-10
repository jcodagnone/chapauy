{ pkgs, inputs, lib, ... }:

{
  stdenv = pkgs.stdenvNoCC;

  packages = with pkgs; [
    # ── Runtimes ─────────────────────────────────────────────────────────
    go           # main module (go.mod) + Dagger module (.dagger/go.mod)
    nodejs_24    # web/ frontend
    pnpm         # package manager for web/

    # ── Build / container orchestration ──────────────────────────────────
    inputs.dagger.packages.${pkgs.system}.dagger  # pinned from github:dagger/nix
    podman

    # ── Go linting & static analysis ─────────────────────────────────────
    golangci-lint   # make lint
    gosec           # make security
    govulncheck     # make vuln
    go-tools        # staticcheck (honnef.co/go/tools)

    # ── Supply-chain / SBOM ───────────────────────────────────────────────
    syft            # make sbom (spdx)
    cosign          # make sign / verify
    cyclonedx-gomod # make sbom (cyclonedx)

    # ── License headers ───────────────────────────────────────────────────
    addlicense      # make license / addlicense

    # ── General dev utilities ─────────────────────────────────────────────
    gnumake
    git
    jq
    curl
  ];

  env = {
    DAGGER_NO_NAG = "1";
  };

  enterShell = ''
    export GOPATH="''${GOPATH:-$HOME/go}"
    export PATH="$GOPATH/bin:$PATH"

    # bom (sigs.k8s.io/bom) is not packaged in nixpkgs; install once if absent.
    if ! command -v bom &>/dev/null; then
      echo "devenv: installing bom (sigs.k8s.io/bom)…"
      go install sigs.k8s.io/bom/cmd/bom@latest
    fi
  '';
}
