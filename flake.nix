{
  description = "Specture - spec-driven software architecture system";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        pname = "specture";
        version = "0.2.2";
      in
      {
        packages.default = pkgs.buildGoModule {
          inherit pname version;
          src = self;
          # vendorHash locks Go module dependencies
          vendorHash = "sha256-FnNfDAK60gj9+KPJsbDgy6ZOUlJu67ywf2WtrNY8sUs=";
          doCheck = false;
          meta = with pkgs.lib; {
            description = "Spec-driven software architecture system";
            homepage = "https://github.com/specture-system/specture";
            license = licenses.mit;
            platforms = platforms.all;
          };
        };

        packages.specture = self.packages.${system}.default;

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/specture";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            git
            just
            pre-commit
          ];
        };
      }
    );
}
