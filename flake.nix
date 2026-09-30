{
  description = "A Go library and CLI for Nextcloud Tables";

  inputs = {
    nixpkgs = {
      url = "https://nixos.org/channels/nixpkgs-unstable/nixexprs.tar.xz";
    };
  };

  outputs =
    {
      nixpkgs,
      ...
    }:
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
                  localPackages = {
                    tables = final.callPackage ./tables.nix { };
                  };
                })
              ];
            };
            inherit system;
          }
        );

      dependencies = pkgs: [
        pkgs.coreutils
        pkgs.docker
        pkgs.gawk
        pkgs.gcc
        pkgs.git
        pkgs.go
        pkgs.go-task
        pkgs.ginkgo
        pkgs.golangci-lint
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
          ciEnvironment = pkgs.mkShellNoCC {
            packages = (dependencies pkgs) ++ (ciDependencies pkgs);
          };

          cdEnvironment = pkgs.mkShellNoCC {
            packages = (dependencies pkgs) ++ (cdDependencies pkgs);
          };

          devEnvironment = pkgs.mkShellNoCC {
            packages =
              (dependencies pkgs) ++ (ciDependencies pkgs) ++ (cdDependencies pkgs) ++ (devDependencies pkgs);
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
              (dependencies pkgs) ++ (ciDependencies pkgs) ++ (cdDependencies pkgs) ++ (devDependencies pkgs);
          };
        }
      );
    };
}
