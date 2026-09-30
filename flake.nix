{
  description = "Nextcloud Tables Go library and CLI";

  inputs = {
    nixpkgs.url = "https://nixos.org/channels/nixpkgs-unstable/nixexprs.tar.xz";
  };

  outputs =
    { nixpkgs, ... }:
    let
      supportedSystems = nixpkgs.lib.systems.flakeExposed;
      forAllSystems =
        function:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          function {
            pkgs = import nixpkgs {
              inherit system;
              overlays = [
                (final: previous: {
                  localPackages.tables = final.callPackage ./tables.nix { };
                })
              ];
            };
            inherit system;
          }
        );
      dependencies = pkgs: [
        pkgs.git
        pkgs.go
        pkgs.go-task
        pkgs.golangci-lint
        pkgs.ginkgo
        pkgs.gawk
        pkgs.coreutils
        pkgs.docker
        pkgs.nix
      ];
      devDependencies = pkgs: [
        pkgs.nixd
        pkgs.nixfmt
      ];
      ciDependencies = pkgs: [ ];
      cdDependencies = pkgs: [ ];
      fmtDependencies = pkgs: pkgs.nixfmt-tree;
    in
    {
      formatter = forAllSystems ({ pkgs, ... }: fmtDependencies pkgs);

      devShells = forAllSystems (
        { pkgs, ... }:
        {
          ciEnvironment = pkgs.mkShell {
            packages = (dependencies pkgs) ++ (ciDependencies pkgs);
          };
          cdEnvironment = pkgs.mkShell {
            packages = (dependencies pkgs) ++ (cdDependencies pkgs);
          };
          devEnvironment = pkgs.mkShell {
            packages =
              (dependencies pkgs)
              ++ (ciDependencies pkgs)
              ++ (cdDependencies pkgs)
              ++ (devDependencies pkgs);
          };
        }
      );

      packages = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.localPackages.tables;
          tables = pkgs.localPackages.tables;
          devEnvironment = pkgs.buildEnv {
            name = "development environment";
            paths =
              (dependencies pkgs)
              ++ (ciDependencies pkgs)
              ++ (cdDependencies pkgs)
              ++ (devDependencies pkgs);
          };
        }
      );
    };
}
