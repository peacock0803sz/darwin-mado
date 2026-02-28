{ lib, pkgs, config, ... }:
let
  cfg = config.programs.mado;
  generatedOptions = import ./generated-options.nix { inherit lib; };
  yamlFormat = pkgs.formats.yaml { };

  # Prune all null values recursively before YAML serialization (R-007)
  filterNulls = lib.filterAttrsRecursive (_: v: v != null);

  # Detect whether the user explicitly configured any settings (FR-009)
  hasSettings = lib.filterAttrs (_: v: v != null) cfg.settings != { };
in
{
  options.programs.mado = {
    enable = lib.mkEnableOption "mado window manager CLI";

    package = lib.mkOption {
      type = lib.types.package;
      description = "The mado package to install. Pass from flake input: mado.packages.\${system}.default";
    };

    settings = lib.mkOption {
      type = lib.types.submodule {
        options = generatedOptions.options;
      };
      default = { };
      description = "mado configuration. All settings default to null (not written to config file).";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    # Config is written to /etc/mado/config.yaml.
    # The CLI discovers it via its fallback search order (FR-012).
    # No $MADO_CONFIG env var is set — user-level HM config at
    # ~/.config/mado/config.yaml takes precedence over this system-level file.
    environment.etc."mado/config.yaml" = lib.mkIf hasSettings {
      source = yamlFormat.generate "mado-config" (filterNulls cfg.settings);
    };

    assertions = generatedOptions.assertions cfg;
  };
}
