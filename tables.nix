{
  buildGoModule,
  lib,
}:
buildGoModule {
  pname = "tables";
  version = "0.1.0";

  src = ./.;

  vendorHash = lib.fakeHash;

  subPackages = [ "cmd/tables" ];

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Go library and CLI for Nextcloud Tables";
    mainProgram = "tables";
  };
}
