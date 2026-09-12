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

  vendorHash = "sha256-RGfYVhSFb/fFrLJ6PwG4ZXLuHdWcdYm2LmLLU7flM/s=";

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
