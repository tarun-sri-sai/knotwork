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
        perSystem = { lib, pkgs, ... }: {
          packages = rec {
            knotwork = pkgs.buildGoModule {
              pname = "knotwork";
              version = "0.1.0";

              src = ./core;
              vendorHash = "sha256-7AaW3amqIzITMgIFrKoqc5JJkVDcsGqFynL38Sk0fB4=";

              postInstall = ''
                mv $out/bin/app $out/bin/knotwork
              '';
            };

            knotwork-mcp = pkgs.buildGoModule {
              pname = "knotwork-mcp";
              version = "0.1.0";

              src = ./mcp;
              vendorHash = "sha256-hbV/kOuImCWwmxcOdg9bEM8VLrCcn0m5bF4MMmj9lSs=";

              postInstall = ''
                mv $out/bin/app $out/bin/knotwork-mcp
              '';
            };

            default = knotwork;
          };

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
