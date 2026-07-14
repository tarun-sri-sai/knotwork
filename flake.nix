{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } (
      top@{ ... }:
      {
        systems = [ "x86_64-linux" ];
        perSystem = { pkgs, ... }: {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              nixd
              nixfmt
            ];

            buildInputs = with pkgs; [
              go
              gotools
            ];

            shellHook = ''
              export GOPATH="$PWD/.go"
              export GOMODCACHE="$GOPATH/pkg/mod"
              export GOCACHE="$GOPATH/cache"
              export GOBIN="$GOPATH/bin"

              mkdir -p "$GOMODCACHE" "$GOCACHE" "$GOBIN"
            '';
          };
        };
      }
    );

}
