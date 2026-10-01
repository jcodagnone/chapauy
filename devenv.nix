# Copyright 2026 The ChapaUY Authors
# SPDX-License-Identifier: Apache-2.0

{ pkgs, inputs, ... }:

let
  # web/package.json pins this exact pnpm. nixpkgs ships a newer major, which
  # rewrites pnpm-lock.yaml. The npm tarball is the same artifact Corepack
  # would download, checked by hash.
  pnpm = pkgs.stdenvNoCC.mkDerivation (finalAttrs: {
    pname = "pnpm";
    version = "10.33.3";

    src = pkgs.fetchurl {
      url = "https://registry.npmjs.org/pnpm/-/pnpm-${finalAttrs.version}.tgz";
      hash = "sha256-PiqQYxIqm0mRszbKkBY8gguOisX0UPs8mwvVaU+UNm8=";
    };

    sourceRoot = "package";
    nativeBuildInputs = [ pkgs.makeWrapper ];
    dontBuild = true;

    installPhase = ''
      runHook preInstall
      mkdir -p $out/libexec/pnpm
      cp -R . $out/libexec/pnpm
      makeWrapper ${pkgs.nodejs_24}/bin/node $out/bin/pnpm \
        --add-flags $out/libexec/pnpm/bin/pnpm.cjs
      makeWrapper ${pkgs.nodejs_24}/bin/node $out/bin/pnpx \
        --add-flags $out/libexec/pnpm/bin/pnpx.cjs
      runHook postInstall
    '';
  });

  # Published by github:dagger/nix. dagger.json selects the engine the CLI runs.
  dagger = inputs.dagger.packages.${pkgs.stdenv.hostPlatform.system}.dagger;
in
{
  # Default stdenv (not stdenvNoCC): duckdb-go links a bundled static
  # libduckdb with cgo and needs gcc plus libstdc++.
  packages = with pkgs; [
    # Runtimes
    go
    nodejs_24
    pnpm
    python3 # node-gyp, if the duckdb npm package falls back to a source build

    # Data, containers, cloud
    duckdb
    dagger
    podman # wrapped with crun, conmon, netavark, aardvark-dns, passt
    google-cloud-sdk

    # Go lint and supply chain, the tools `make` invokes
    golangci-lint
    gosec
    govulncheck
    go-tools # staticcheck
    syft
    cosign
    cyclonedx-gomod
    addlicense

    gnumake
    git
    jq
    curl
  ];

  env.DAGGER_NO_NAG = "1";
  env.CGO_ENABLED = "1";

  enterShell = ''
    if [ -n "''${HOME:-}" ]; then
      export GOPATH="''${GOPATH:-$HOME/go}"
      case ":$PATH:" in
        *":$GOPATH/bin:"*) ;;
        *) export PATH="$PATH:$GOPATH/bin" ;;
      esac

      # sigs.k8s.io/bom is not packaged in nixpkgs. `make deps` installs it too.
      if ! command -v bom >/dev/null 2>&1; then
        echo "devenv: instalando bom (sigs.k8s.io/bom) en \$GOPATH/bin…"
        go install sigs.k8s.io/bom/cmd/bom@latest \
          || echo "devenv: no se pudo instalar bom."
      fi
    fi
  '';
}
