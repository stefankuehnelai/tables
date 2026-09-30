{
  lib,
  buildGoModule,
}:

buildGoModule rec {
  pname = "tables";
  version = "0.1.0";

  src = builtins.path {
    path = ./.;
    name = "tables";
  };

  # Replace this with the hash reported by the first reproducible Nix build.
  vendorHash = lib.fakeHash;

  env.CGO_ENABLED = 0;

  subPackages = [ "cmd/tables" ];
  doCheck = false;

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Go library and CLI for Nextcloud Tables";
    homepage = "https://github.com/stefankuehnelai/tables";
    license = licenses.gpl3Only;
    mainProgram = "tables";
  };
}
