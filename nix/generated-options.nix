# AUTO-GENERATED from schemas/config.v1.schema.json
# Do not edit manually. Run: go run ./cmd/nix-options-gen -schema schemas/config.v1.schema.json -output nix/generated-options.nix
{ lib }:
{
  options = {
    format = lib.mkOption {
      type = lib.types.nullOr (lib.types.enum [ "text" "json" ]);
      default = null;
      description = "Default output format";
    };
    ignore_apps = lib.mkOption {
      type = lib.types.nullOr (lib.types.listOf (lib.types.str));
      default = null;
      description = "Application names to exclude from list output and preset matching";
    };
    presets = lib.mkOption {
      type = lib.types.nullOr (lib.types.listOf (lib.types.submodule {
        options = {
          description = lib.mkOption {
            type = lib.types.nullOr (lib.types.str);
            default = null;
            description = "Human-readable description of the preset";
          };
          name = lib.mkOption {
            type = lib.types.str;
            description = "Preset name (alphanumeric, hyphens, underscores)";
          };
          rules = lib.mkOption {
            type = lib.types.listOf (lib.types.submodule {
              options = {
                app = lib.mkOption {
                  type = lib.types.str;
                  description = "Application name (case-insensitive exact match)";
                };
                desktop = lib.mkOption {
                  type = lib.types.nullOr (lib.types.ints.unsigned);
                  default = null;
                  description = "Desktop number to scope this rule to (0 = windows assigned to all desktops)";
                };
                position = lib.mkOption {
                  type = lib.types.nullOr (lib.types.listOf (lib.types.int));
                  default = null;
                  description = "Target position [x, y] in global coordinates";
                };
                screen = lib.mkOption {
                  type = lib.types.nullOr (lib.types.str);
                  default = null;
                  description = "Screen ID or name filter";
                };
                size = lib.mkOption {
                  type = lib.types.nullOr (lib.types.listOf (lib.types.ints.positive));
                  default = null;
                  description = "Target size [width, height] (positive integers)";
                };
                title = lib.mkOption {
                  type = lib.types.nullOr (lib.types.str);
                  default = null;
                  description = "Window title filter (case-insensitive partial match)";
                };
              };
            });
            description = "Window operation rules (evaluated in order, first match wins)";
          };
        };
      }));
      default = null;
      description = "Named window layout presets";
    };
    timeout = lib.mkOption {
      type = lib.types.nullOr (lib.types.str);
      default = null;
      description = "AX operation timeout (e.g. 5s, 10s, 1m)";
    };
  };
  assertions = cfg: [
    {
      assertion = cfg.settings.ignore_apps == null || builtins.all (s: builtins.stringLength s >= 1) cfg.settings.ignore_apps;
      message = "ignore_apps items must be non-empty strings";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.match "^[a-zA-Z0-9][a-zA-Z0-9_-]*$" p.name != null) cfg.settings.presets;
      message = "name must match pattern ^[a-zA-Z0-9][a-zA-Z0-9_-]*$";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.length p.rules >= 1) cfg.settings.presets;
      message = "rules must have at least 1 item(s)";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.all (r: (r.position != null) || (r.size != null)) p.rules) cfg.settings.presets;
      message = "Each preset rule must have at least 'position' or 'size'";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.all (r: r.position == null || builtins.length r.position >= 2) p.rules) cfg.settings.presets;
      message = "position must have at least 2 item(s)";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.all (r: r.position == null || builtins.length r.position <= 2) p.rules) cfg.settings.presets;
      message = "position must have at most 2 item(s)";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.all (r: r.size == null || builtins.length r.size >= 2) p.rules) cfg.settings.presets;
      message = "size must have at least 2 item(s)";
    }
    {
      assertion = cfg.settings.presets == null || builtins.all (p: builtins.all (r: r.size == null || builtins.length r.size <= 2) p.rules) cfg.settings.presets;
      message = "size must have at most 2 item(s)";
    }
    {
      assertion = cfg.settings.timeout == null || builtins.match "^[0-9]+(ns|us|ms|s|m|h)$" cfg.settings.timeout != null;
      message = "timeout must match pattern ^[0-9]+(ns|us|ms|s|m|h)$";
    }
  ];
}
