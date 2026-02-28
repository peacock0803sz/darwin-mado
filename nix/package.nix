{
  lib,
  buildGoModule,
}:
let
  version = "0.0.0-dev";
in
buildGoModule {
  pname = "mado";
  inherit version;

  src = lib.cleanSource ../.;

  vendorHash = "sha256-y8ZUtc70LFItESZsLtor/pd7vJusvCH4AwYzAl0y8u0=";

  env.CGO_ENABLED = "1";

  ldflags = [
    "-X main.version=${version}"
  ];

  subPackages = [ "cmd/mado" ];

  meta = {
    description = "macOS window manager CLI";
    homepage = "https://github.com/peacock0803sz/darwin-mado";
    platforms = lib.platforms.darwin;
    mainProgram = "mado";
  };
}
