{
  description = "";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.git-hooks.flakeModule ];
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];

      flake = {
        homeManagerModules.default = ./nix/hm-module.nix;
        darwinModules.default = ./nix/darwin-module.nix;
      };
      perSystem = { config, pkgs, lib, ... }: {
        packages = lib.optionalAttrs pkgs.stdenv.isDarwin {
          default = pkgs.callPackage ./nix/package.nix { };
        };

        pre-commit.settings.hooks = {
          golangci-lint.enable = true;
          gofumpt = {
            enable = true;
            entry = "${pkgs.gofumpt}/bin/gofumpt -w";
            types = [ "go" ];
          };
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ config.pre-commit.devShell ];
          packages = with pkgs; [
            git
            go
            golangci-lint
            gofumpt
          ];
        };
      };
    };
}
